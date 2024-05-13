package spidey_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ysoding/spidey/cmd"
	"github.com/ysoding/spidey/spidey"
)

var validPaths = map[string]bool{
	"/bootstrap/css/bootstrap.css":                              true,
	"/assets/application-4b77637cc302ef4af6c358864df26f88.css":  true,
	"https://www.youtube.com/player_api":                        true,
	"/assets/application-9709f2e1ad6d5ec24402f59507f6822b.js":   true,
	"/assets/application-valum.js.js":                           true,
	"/services":                                                 true,
	"/contacts":                                                 true,
	"/assets/ardan-symbol-93ee488d16f9bc56ad65659c2d8f41dc.png": true,
	"/assets/member1-55a2b7ac0a868d49fdf50ce39f0ce1ac.png":      true,
	"/assets/member2-66485427ca4bd140e0547efb1ce12ce0.png":      true,
	"/assets/member4-cfa03a1a15aed816528b8ec1ee6c95c6.png":      true,
	"/assets/member5-6ee6a979c39c81e2b652f268cccaf265.png":      true,
}

var events cmd.Events

var index = []byte(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
	<title>Ardan Studios</title>
</head>
<body>

		<link rel="stylesheet" href="/bootstrap/css/bootstrap.css">
		<link href="/assets/application-4b77637cc302ef4af6c358864df26f88.css" media="screen" rel="stylesheet" />

		<script src="https://www.youtube.com/player_api"></script>
		<script src="/assets/application-9709f2e1ad6d5ec24402f59507f6822b.js"></script>
		<script src="/assets/application-valum.js.js"></script>

		<a href="/services"></a>
		<a href="/contacts"></a>

		<a href="http://youtube.com/x8433j4i"></a>
		<a href="http://gracehound.com/index"></a>

		<img class="ardan-symbol" src="/assets/ardan-symbol-93ee488d16f9bc56ad65659c2d8f41dc.png" />
		<img src="/assets/member1-55a2b7ac0a868d49fdf50ce39f0ce1ac.png" />
		<img src="/assets/member2-66485427ca4bd140e0547efb1ce12ce0.png" />
		<img src="/assets/member4-cfa03a1a15aed816528b8ec1ee6c95c6.png" />
		<img src="/assets/member5-6ee6a979c39c81e2b652f268cccaf265.png" />

</body>
</html>`)

var badImages = []byte(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
	<title>Bad Images</title>
<body>

		<img class="ardan-symbol" src="/assets/ardan-symbol-93ee488d16f9bc56ad65659c2d8f41dc.png" />
		<img src="/assets/member1-55a2b7ac0a868d49fdf50ce39f0ce1ac.png" />
		<img src="/assets/member2-66485427ca4bd140e0547efb1ce12ce0.png" />
		<img src="/assets/member4-cfa03a1a15aed816528b8ec1ee6c95c6.png" />
		<img src="/assets/member5-6ee6a979c39c81e2b652f268cccaf265.png" />
		<img src="/assets/member6-e202d0df26e17043328648feda1fc327.png" />
		<img src="/assets/member7-cbcbc8bfe0d8f0cefe66a1b801827f74.png" />
		<img src="/assets/member10-01462a64b08492ba3b64058ea50b94f8.png" />

</body>
</html>`)

var badLinks = []byte(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Bad Links</title>
</head>
<body>

		<link rel="stylesheet" href="/bootstrap/css/bootstrap.css">
		<link href="/maxcdn.bootstrapcdn.com/font-awesome/4.2.0/css/font-awesome.min.css" type="text/css" rel="stylesheet" />
		<link href="/assets/application-4b77637cc302ef4af6c358864df26f88.css" media="screen" rel="stylesheet" />

		<a href="/services"></a>
		<a href="/contacts"></a>
		<a href="/billy"></a>
		<a href="/wacksee"></a>

		<a href="http://youtube.com/x8433j4i"></a>
		<a href="http://gracehound.com/index"></a>

</body>
</html>`)

var badScripts = []byte(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
	<title>Bad Scripts</title>
</head>
<body>
	<script src="/assets/application-9709f2e1ad6d5ec24402f59507f6822b.js"></script>
	<script src="/assets/application-blacksmith.js"></script>
	<script src="/assets/application-trottle.js.js"></script>
	<script src="/assets/application-valum.js.js"></script>
	<script src="https://www.youtube.com/player_api"></script>
</body>
</html>`)

func Test_Spidey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			if !validPaths[req.URL.Path] {
				res.WriteHeader(http.StatusNotFound)
				return
			}
		}

		req.ParseForm()

		page := strings.TrimSpace(req.FormValue("page"))

		switch page {
		case "badscripts":
			res.Write(badScripts)
			return
		case "badlinks":
			res.Write(badLinks)
			return
		case "badimages":
			res.Write(badImages)
			return
		}
		res.WriteHeader(200)
		res.Write(index)
	}))

	defer server.Close()
	conf := spidey.Config{
		Client:              &http.Client{Timeout: 6 * time.Second},
		URL:                 server.URL,
		Depth:               3,
		Events:              events,
		EnableCheckExternal: false,
	}

	testValidIndex(conf, t)
	testInvalidScripts(conf, t)
	testInvalidLinks(conf, t)
	testInvalidImages(conf, t)

}

func testValidIndex(c spidey.Config, t *testing.T) {
	res, err := spidey.Run(context.Background(), &c)
	if err != nil {
		t.Fatalf("Should have successfully retrieved page[%s]: %q", c.URL, err)
	}

	if len(res.DeadLinks) != 0 {
		t.Fatalf("Should have found no deadlinks in page[%s]: %+v", c.URL, res.DeadLinks)
	}
}

func testInvalidImages(c spidey.Config, t *testing.T) {
	c.URL = fmt.Sprintf("%s?page=badimages", c.URL)
	res, err := spidey.Run(context.Background(), &c)
	if err != nil {
		t.Fatalf("Should have successfully retrieved page[%s]: %q", c.URL, err)
	}

	if len(res.DeadLinks) != 3 {
		t.Fatalf("Should have found 3 dead image links in page[%s]: %+v", c.URL, res.DeadLinks)
	}
}

func testInvalidScripts(c spidey.Config, t *testing.T) {
	c.URL = fmt.Sprintf("%s?page=badscripts", c.URL)
	res, err := spidey.Run(context.Background(), &c)
	if err != nil {
		t.Fatalf("Should have successfully retrieved page[%s]: %q", c.URL, err)
	}

	if len(res.DeadLinks) != 2 {
		t.Fatalf("Should have found 2 dead script links in page[%s]: %+v", c.URL, res.DeadLinks)
	}
}

func testInvalidLinks(c spidey.Config, t *testing.T) {
	c.URL = fmt.Sprintf("%s?page=badlinks", c.URL)
	res, err := spidey.Run(context.Background(), &c)
	if err != nil {
		t.Fatalf("Should have successfully retrieved page[%s]: %q", c.URL, err)
	}

	if len(res.DeadLinks) != 3 {
		t.Fatalf("Should have found 3 dead links in page[%s]: %+v", c.URL, res.DeadLinks)
	}
}
