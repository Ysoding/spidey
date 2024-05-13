package spidey

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ysoding/spidey/pool"
	"golang.org/x/net/html/charset"
)

type LinkReport struct {
	Link   string
	Status int
	Error  error
}

type SpideyResult struct {
	DeadLinks []LinkReport
	Start     time.Time
	End       time.Time
}

func (sr *SpideyResult) ResultFormat() string {
	var sb strings.Builder

	sb.WriteString("\n\n--------------Timelapse---------------\n")
	sb.WriteString(fmt.Sprintf(`
	Start Time: %s
	End Time: %s
	Duration: %s
	`, sr.Start, sr.End, sr.End.Sub(sr.Start)))

	if len(sr.DeadLinks) > 0 {
		sb.WriteString("\n\n--------------DEAD LINKS--------------\n")

		for _, lr := range sr.DeadLinks {
			sb.WriteString(fmt.Sprintf(`
		URL: %s
		Status Code: %d
		Error: %s
		`, lr.Link, lr.Status, lr.Error))
		}
	} else {
		sb.WriteString("\n\n--------------NO DEAD LINKS--------------\n")
	}

	return sb.String()
}

type Events interface {
	Event(context interface{}, event string, format string, data ...interface{})
	ErrorEvent(context interface{}, event string, err error, format string, data ...interface{})
}

func Run(context interface{}, c *Config) (*SpideyResult, error) {
	c.Events.Event(context, "Run", "Started: URL[%s]", c.URL)
	res := &SpideyResult{DeadLinks: make([]LinkReport, 0)}

	path, err := url.Parse(c.URL)
	if err != nil {
		c.Events.ErrorEvent(context, "Run", err, "Completed")
		return nil, err
	}

	deadCh := make(chan LinkReport)

	res.Start = time.Now()
	go start(c, path, deadCh)

	for lp := range deadCh {
		res.DeadLinks = append(res.DeadLinks, lp)
	}

	res.End = time.Now().UTC()
	c.Events.Event(context, "Run", "Completed")
	return res, nil
}

func start(c *Config, path *url.URL, deadCh chan LinkReport) {
	visited := make(map[string]bool)
	var wait sync.WaitGroup
	var mu sync.RWMutex
	pool := pool.New()

	wait.Add(1)
	pool.Do("start", &pathBot{
		path:    path.String(),
		deadCh:  deadCh,
		config:  c,
		wait:    &wait,
		visited: visited,
		mu:      &mu,
		index:   path,
		pool:    pool,
		depth:   0,
	})
	wait.Wait()
	close(deadCh)
	pool.Shutdown()
}

type pathBot struct {
	path    string
	deadCh  chan LinkReport
	config  *Config
	wait    *sync.WaitGroup
	visited map[string]bool
	mu      *sync.RWMutex
	index   *url.URL
	pool    *pool.Pool
	depth   int
}

func (p pathBot) Work(context interface{}) {
	defer p.wait.Done()

	if p.depth > p.config.Depth {
		return
	}

	p.mu.RLock()
	exists := p.visited[p.path]
	p.mu.RUnlock()

	if exists {
		return
	}

	p.mu.Lock()
	p.visited[p.path] = true
	p.mu.Unlock()

	p.config.Events.Event(context, "Check", "URL[%s] Depth[%d] Start ", p.path, p.depth)

	status, crawleable, err := checkPath(p.path, p.config)
	if err != nil {
		p.deadCh <- LinkReport{Link: p.path, Status: status, Error: err}
		return
	}

	if !crawleable {
		p.config.Events.Event(context, "URL Status", "URL[%s] not craweable", p.path)
		return
	}

	links := make(chan string)
	if err := getAllLinks(p.path, links); err != nil {
		p.deadCh <- LinkReport{Link: p.path, Status: http.StatusInternalServerError, Error: err}
		return
	}

	for link := range links {
		p.config.Events.Event(context, "Found Link", "Link[%s]", link)

		p.mu.RLock()
		visited := p.visited[link]
		p.mu.RUnlock()

		if visited {
			continue
		}

		if strings.HasPrefix(link, "#") {
			p.mu.Lock()
			p.visited[link] = true
			p.mu.Unlock()
			continue
		}

		if strings.TrimSpace(link) == "/" || link == p.path {
			continue
		}

		pathURI, err := parsePath(link, p.index)
		if err != nil {
			continue
		}

		p.mu.RLock()
		visited = p.visited[pathURI.Path]
		p.mu.RUnlock()
		if visited {
			continue
		}

		if !p.config.EnableCheckExternal && !strings.Contains(pathURI.Host, p.index.Host) {
			p.mu.Lock()
			p.visited[link] = true
			p.visited[pathURI.Path] = true
			p.mu.Unlock()
			continue
		}

		p.mu.Lock()
		p.visited[link] = true
		p.visited[pathURI.Path] = true
		p.mu.Unlock()

		status, crawlable, err := checkPath(pathURI.String(), p.config)
		if err != nil {
			p.deadCh <- LinkReport{Link: p.path, Status: status, Error: err}
		}

		if !crawlable {
			continue
		}

		p.wait.Add(1)
		p.pool.Do(context, &pathBot{
			path:    pathURI.String(),
			deadCh:  p.deadCh,
			config:  p.config,
			wait:    p.wait,
			index:   p.index,
			visited: p.visited,
			mu:      p.mu,
			pool:    p.pool,
			depth:   p.depth + 1,
		})
	}
	p.config.Events.Event(context, "Check", "URL[%s] Depth[%d] End", p.path, p.depth)
}

// checkPath check link status, only crawl html web page
func checkPath(path string, c *Config) (status int, shouldCrawl bool, err error) {
	res, err := c.Client.Head(path)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	status = res.StatusCode
	if res.StatusCode < 200 || res.StatusCode > 299 {
		err = errors.New("link failed")
		return
	}

	if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
		return
	}

	shouldCrawl = true
	return
}

func parsePath(link string, index *url.URL) (*url.URL, error) {
	res, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if !res.IsAbs() {
		res = index.ResolveReference(res)
	}
	return res, nil
}

func getAllLinks(link string, port chan string) error {
	// src href
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	bodyReader, err := charset.NewReader(resp.Body, ct)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return err
	}

	go func() {
		defer close(port)

		doc.Find("[src]").Each(func(_ int, s *goquery.Selection) {
			src, exists := s.Attr("src")
			if !exists {
				return
			}
			if strings.Contains(src, "javascript:void(0)") {
				return
			}
			port <- src
		})

		doc.Find("[href]").Each(func(_ int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}
			if strings.Contains(href, "javascript:void(0)") {
				return
			}
			port <- href
		})

	}()

	return nil
}
