package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Downloader struct {
	BaseURL   *url.URL
	RootDir   string
	MaxDepth  int
	Client    *http.Client
	Visited   map[string]struct{}
	Mu        sync.Mutex
	Semaphore chan struct{}
	Errors    []error
	ErrMu     sync.Mutex
	wg        sync.WaitGroup
}

func NewDownloader(rawurl, rootDir string, maxDepth, parallel int) (*Downloader, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	return &Downloader{
		BaseURL:   u,
		RootDir:   rootDir,
		MaxDepth:  maxDepth,
		Client:    &http.Client{Timeout: 15 * time.Second},
		Visited:   make(map[string]struct{}),
		Semaphore: make(chan struct{}, parallel),
		Errors:    []error{},
	}, nil
}

func (d *Downloader) appendError(err error) {
	d.ErrMu.Lock()
	d.Errors = append(d.Errors, err)
	d.ErrMu.Unlock()
}

func (d *Downloader) Start() {
	d.download(d.BaseURL.String(), 0)
	d.wg.Wait()
}

func (d *Downloader) download(rawurl string, depth int) {
	if depth > d.MaxDepth {
		return
	}

	normalized, err := d.normalizeURL(rawurl)
	if err != nil {
		return
	}

	d.Mu.Lock()
	if _, ok := d.Visited[normalized]; ok {
		d.Mu.Unlock()
		return
	}
	d.Visited[normalized] = struct{}{}
	d.Mu.Unlock()

	d.wg.Add(1)
	d.Semaphore <- struct{}{}

	go func(url string, currentDepth int) {
		defer d.wg.Done()
		defer func() { <-d.Semaphore }()

		fmt.Printf("Downloading %s (depth %d)\n", url, currentDepth)
		resp, err := d.Client.Get(url)
		if err != nil {
			d.appendError(fmt.Errorf("error fetching %s: %v", url, err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			d.appendError(fmt.Errorf("bad status for %s: %s", url, resp.Status))
			return
		}

		contentType := resp.Header.Get("Content-Type")
		isHTML := strings.Contains(contentType, "text/html")

		localPath := d.urlToFilePath(url, isHTML)

		err = os.MkdirAll(filepath.Dir(localPath), 0755)
		if err != nil {
			d.appendError(fmt.Errorf("error creating dir for %s: %v", localPath, err))
			return
		}

		f, err := os.Create(localPath)
		if err != nil {
			d.appendError(fmt.Errorf("error creating file %s: %v", localPath, err))
			return
		}
		defer f.Close()

		if isHTML {
			doc, err := html.Parse(resp.Body)
			if err != nil {
				d.appendError(fmt.Errorf("html parse error %s: %v", url, err))
				io.Copy(f, resp.Body)
				return
			}
			d.rewriteLinks(doc)
			html.Render(f, doc)
			f.Sync()

			links := d.collectLinks(doc)
			for _, link := range links {
				normalizedLink, err := d.normalizeURL(link)
				if err != nil || normalizedLink == "" {
					continue
				}
				if d.isInDomain(normalizedLink) {
					d.download(normalizedLink, currentDepth+1)
				}
			}
		} else {
			_, err = io.Copy(f, resp.Body)
			if err != nil {
				d.appendError(fmt.Errorf("error saving resource %s: %v", url, err))
				return
			}
		}
	}(normalized, depth)
}

func (d *Downloader) normalizeURL(rawurl string) (string, error) {
	rawurl = strings.TrimSpace(rawurl)
	if rawurl == "" {
		return "", fmt.Errorf("empty url")
	}

	if strings.Contains(rawurl, "://") {
		u, err := url.Parse(rawurl)
		if err != nil {
			return "", err
		}
		u.Fragment = ""
		return u.String(), nil
	}

	if strings.Contains(rawurl, ".") && strings.Contains(rawurl, "/") &&
		!strings.HasPrefix(rawurl, "/") && !strings.HasPrefix(rawurl, "./") && !strings.HasPrefix(rawurl, "../") {
		rawurl = "https://" + rawurl
		u, err := url.Parse(rawurl)
		if err != nil {
			return "", err
		}
		u.Fragment = ""
		return u.String(), nil
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	u = d.BaseURL.ResolveReference(u)
	u.Fragment = ""
	return u.String(), nil
}

func (d *Downloader) isInDomain(rawurl string) bool {
	u, err := url.Parse(rawurl)
	if err != nil {
		return false
	}
	return u.Hostname() == d.BaseURL.Hostname()
}

func (d *Downloader) urlToFilePath(rawurl string, isHTML bool) string {
	u, _ := url.Parse(rawurl)
	path := u.Path
	if path == "" || strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	} else if isHTML && !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, ".htm") {
		path += ".html"
	}
	localPath := filepath.Join(d.RootDir, u.Hostname(), path)
	return localPath
}

func (d *Downloader) collectLinks(n *html.Node) []string {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if (attr.Key == "href" || attr.Key == "src") && attr.Val != "" {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return links
}

func (d *Downloader) rewriteLinks(n *html.Node) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if (attr.Key == "href" || attr.Key == "src") && attr.Val != "" {
					absurl, err := d.normalizeURL(attr.Val)
					if err == nil && d.isInDomain(absurl) {
						localPath := d.urlToFilePath(absurl, strings.HasSuffix(absurl, ".html") || strings.HasSuffix(absurl, "/"))
						relPath, err := filepath.Rel(d.RootDir, localPath)
						if err == nil {
							n.Attr[i].Val = relPath
						} else {
							n.Attr[i].Val = localPath
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
}

func main() {
	depth := flag.Int("d", 1, "depth for recursive downloads")
	parallel := flag.Int("n", 3, "number of parallel downloads")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: wgetmirror [-d depth] [-n parallel] URL")
		os.Exit(1)
	}
	url := args[0]
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	rootDir := "mirror"

	downloader, err := NewDownloader(url, rootDir, *depth, *parallel)
	if err != nil {
		fmt.Printf("Error initializing downloader: %v\n", err)
		os.Exit(1)
	}

	downloader.Start()

	downloader.wg.Wait()

	if len(downloader.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "Download completed with errors:")
		for _, e := range downloader.Errors {
			fmt.Fprintf(os.Stderr, "- %v\n", e)
		}
		os.Exit(1)
	}
	fmt.Println("Download complete.")
}
