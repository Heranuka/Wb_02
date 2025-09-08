package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type mockRoundTripper struct {
	responses map[string]*http.Response
	errs      map[string]error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if err, ok := m.errs[req.URL.String()]; ok {
		return nil, err
	}
	if resp, ok := m.responses[req.URL.String()]; ok {
		return resp, nil
	}
	return nil, errors.New("not found")
}

func newHTTPResponse(body string, statusCode int, contentType string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{contentType}},
	}
}

func TestNormalizeURL(t *testing.T) {
	d, _ := NewDownloader("https://example.com/dir/", "mirror", 2, 2)

	cases := []struct {
		input string
		want  string
	}{
		{"about.html", "https://example.com/dir/about.html"},
		{"/home", "https://example.com/home"},
		{"https://other.com/p", "https://other.com/p"},
		{"../up", "https://example.com/up"},
	}

	for _, c := range cases {
		got, err := d.normalizeURL(c.input)
		if err != nil || got != c.want {
			t.Errorf("normalizeURL(%q) = %q, %v; want %q", c.input, got, err, c.want)
		}
	}
}

func TestIsInDomain(t *testing.T) {
	d, _ := NewDownloader("https://example.com/dir/", "mirror", 3, 5)

	cases := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"https://sub.example.com", false},
		{"https://other.com", false},
		{"http://example.com", true},
	}

	for _, c := range cases {
		got := d.isInDomain(c.url)
		if got != c.want {
			t.Errorf("isInDomain(%q) = %v; want %v", c.url, got, c.want)
		}
	}
}

func TestUrlToFilePath(t *testing.T) {
	d, _ := NewDownloader("https://example.com", "mirror", 1, 2)

	cases := []struct {
		url    string
		isHTML bool
		want   string
	}{
		{"https://example.com/", true, "mirror/example.com/index.html"},
		{"https://example.com/abc", true, "mirror/example.com/abc.html"},
		{"https://example.com/img.png", false, "mirror/example.com/img.png"},
		{"https://example.com/dir/", true, "mirror/example.com/dir/index.html"},
	}

	for _, c := range cases {
		got := d.urlToFilePath(c.url, c.isHTML)
		if got != c.want {
			t.Errorf("urlToFilePath(%q,%v) = %q; want %q", c.url, c.isHTML, got, c.want)
		}
	}
}

func TestCollectLinks(t *testing.T) {
	htmlStr := `<html><head><link href="style.css"></head><body>
        <a href="/page1">Page1</a>
        <img src="pic.jpg"/>
        <script src="app.js"></script>
    </body></html>`
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("html.Parse error: %v", err)
	}
	d, _ := NewDownloader("https://example.com", "mirror", 1, 1)
	links := d.collectLinks(doc)

	expected := []string{"style.css", "/page1", "pic.jpg", "app.js"}

	for _, exp := range expected {
		found := false
		for _, l := range links {
			if l == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("collectLinks missing %q", exp)
		}
	}
}

func TestRewriteLinks(t *testing.T) {
	htmlStr := `<html><head><link href="/css/style.css"></head><body>
<a href="/page1">Page1</a><img src="/img/pic.jpg"/></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("html.Parse error: %v", err)
	}
	d, _ := NewDownloader("https://example.com", "mirror", 1, 1)
	d.rewriteLinks(doc)

	var urls []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if (a.Key == "href" || a.Key == "src") && a.Val != "" {
					urls = append(urls, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	for _, u := range urls {
		if strings.HasPrefix(u, "http") {
			t.Errorf("rewriteLinks url unexpectedly absolute: %q", u)
		}
	}
}

func TestDownloadRecursionAndErrorHandling(t *testing.T) {
	base := "https://example.com/"
	d, _ := NewDownloader(base, "mirror", 2, 2)

	responses := map[string]*http.Response{
		"https://example.com/":        newHTTPResponse(`<a href="/page1">p1</a>`, 200, "text/html"),
		"https://example.com/page1":   newHTTPResponse(`<img src="/img.png"/>`, 200, "text/html"),
		"https://example.com/img.png": newHTTPResponse("image data", 200, "image/png"),
	}
	errs := map[string]error{
		"https://example.com/broken": fmt.Errorf("error fetching broken"),
	}

	d.Client.Transport = &mockRoundTripper{
		responses: responses,
		errs:      errs,
	}

	d.Start()

	if len(d.Errors) != 0 {
		t.Errorf("expected no errors, got %d", len(d.Errors))
	}

	d.download(base, 0)
}
