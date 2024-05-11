package spidey

import (
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

	sb.WriteString("--------------Timelapse---------------\n")
	sb.WriteString(fmt.Sprintf(`
	Start Time: %s
	End Time: %s
	Duration: %s
	`, sr.Start, sr.End, sr.End.Sub(sr.Start)))

	if len(sr.DeadLinks) > 0 {
		sb.WriteString("--------------DEAD LINKS--------------\n")

		for _, lr := range sr.DeadLinks {
			sb.WriteString(fmt.Sprintf(`
		URL: %s
		Status Code: %d
		Error: %s
		`, lr.Link, lr.Status, lr.Error))
		}
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
	defer close(deadCh)

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
	})
	wait.Wait()
	pool.Shutdown()
}

type pathBot struct {
	path    string
	deadCh  chan LinkReport
	config  *Config
	wait    *sync.WaitGroup
	visited map[string]bool
	mu      *sync.RWMutex
}

func (p pathBot) Work(context interface{}) {
	defer p.wait.Done()
	p.mu.RLock()
	exists := p.visited[p.path]
	p.mu.RUnlock()

	if exists {
		return
	}

	p.mu.Lock()
	p.visited[p.path] = true
	p.mu.Unlock()
	// TODO:

	links := make(chan string)
	if err := getAllLinks(p.path, links); err != nil {
		p.deadCh <- LinkReport{Link: p.path, Status: http.StatusInternalServerError, Error: err}
		return
	}

	for link := range links {
		p.deadCh <- LinkReport{Link: link, Status: http.StatusOK, Error: nil}
	}
}

func getAllLinks(url string, port chan string) error {
	// src href
	resp, err := http.Get(url)
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
