package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Downloader struct {
	client *http.Client
	outDir string

	rootHost string

	mu      sync.Mutex
	visited map[string]string
}

func NewDownloader(outDir string, timeout time.Duration) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: timeout,
		},
		outDir:  outDir,
		visited: make(map[string]string),
	}
}

func (d *Downloader) Mirror(rawURL string, depth int) error {
	u, err := d.normalizeURL(rawURL)
	if err != nil {
		return err
	}

	d.rootHost = u.Hostname()

	_, err = d.mirrorPage(u, depth)
	return err
}

func (d *Downloader) mirrorPage(u *url.URL, depth int) (string, error) {
	u, err := d.normalizeURL(u.String())
	if err != nil {
		return "", err
	}

	if !d.sameHost(u) {
		return "", nil
	}

	if localPath, ok := d.getVisited(u.String()); ok {
		return localPath, nil
	}

	localPath := d.localPathForPage(u)
	d.setVisited(u.String(), localPath)

	log.Printf("download page: %s -> %s", u.String(), localPath)

	resp, err := d.fetch(u.String())
	if err != nil {
		d.removeVisited(u.String())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.removeVisited(u.String())
		return "", fmt.Errorf("bad status for %s: %s", u.String(), resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if !isHTML(contentType) {
		if err := writeFile(localPath, body); err != nil {
			d.removeVisited(u.String())
			return "", err
		}
		return localPath, nil
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	if err := d.rewriteHTML(doc, u, localPath, depth); err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	if err := writeFile(localPath, buf.Bytes()); err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	return localPath, nil
}

func (d *Downloader) rewriteHTML(doc *html.Node, pageURL *url.URL, pageLocalPath string, depth int) error {
	var walk func(*html.Node) error

	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				href, ok := getAttr(n, "href")
				if ok {
					newHref, err := d.handlePageLink(pageURL, pageLocalPath, href, depth)
					if err != nil {
						log.Printf("warning: failed to process page link %q: %v", href, err)
					} else if newHref != "" {
						setAttr(n, "href", newHref)
					}
				}

			case "img":
				src, ok := getAttr(n, "src")
				if ok {
					newSrc, err := d.handleAssetLink(pageURL, pageLocalPath, src)
					if err != nil {
						log.Printf("warning: failed to process img src %q: %v", src, err)
					} else if newSrc != "" {
						setAttr(n, "src", newSrc)
					}
				}

			case "script":
				src, ok := getAttr(n, "src")
				if ok {
					newSrc, err := d.handleAssetLink(pageURL, pageLocalPath, src)
					if err != nil {
						log.Printf("warning: failed to process script src %q: %v", src, err)
					} else if newSrc != "" {
						setAttr(n, "src", newSrc)
					}
				}

			case "link":
				href, ok := getAttr(n, "href")
				if ok {
					newHref, err := d.handleAssetLink(pageURL, pageLocalPath, href)
					if err != nil {
						log.Printf("warning: failed to process link href %q: %v", href, err)
					} else if newHref != "" {
						setAttr(n, "href", newHref)
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}

		return nil
	}

	return walk(doc)
}

func (d *Downloader) handlePageLink(pageURL *url.URL, pageLocalPath, rawHref string, depth int) (string, error) {
	if rawHref == "" || strings.HasPrefix(rawHref, "#") {
		return "", nil
	}

	resolved, err := d.resolveURL(pageURL, rawHref)
	if err != nil {
		return "", err
	}

	if shouldSkipURL(resolved) {
		return "", nil
	}

	if !d.sameHost(resolved) {
		return "", nil
	}

	if depth <= 0 {
		return "", nil
	}

	localPath, err := d.mirrorPage(resolved, depth-1)
	if err != nil {
		return "", err
	}

	return relativePath(pageLocalPath, localPath), nil
}

func (d *Downloader) handleAssetLink(pageURL *url.URL, pageLocalPath, rawRef string) (string, error) {
	if rawRef == "" || strings.HasPrefix(rawRef, "#") {
		return "", nil
	}

	resolved, err := d.resolveURL(pageURL, rawRef)
	if err != nil {
		return "", err
	}

	if shouldSkipURL(resolved) {
		return "", nil
	}

	if !d.sameHost(resolved) {
		return "", nil
	}

	localPath, err := d.downloadAsset(resolved)
	if err != nil {
		return "", err
	}

	return relativePath(pageLocalPath, localPath), nil
}

func (d *Downloader) downloadAsset(u *url.URL) (string, error) {
	u, err := d.normalizeURL(u.String())
	if err != nil {
		return "", err
	}

	if localPath, ok := d.getVisited(u.String()); ok {
		return localPath, nil
	}

	localPath := d.localPathForAsset(u)
	d.setVisited(u.String(), localPath)

	log.Printf("download asset: %s -> %s", u.String(), localPath)

	resp, err := d.fetch(u.String())
	if err != nil {
		d.removeVisited(u.String())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.removeVisited(u.String())
		return "", fmt.Errorf("bad status for asset %s: %s", u.String(), resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	if err := writeFile(localPath, body); err != nil {
		d.removeVisited(u.String())
		return "", err
	}

	return localPath, nil
}

func (d *Downloader) fetch(rawURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "mini-wget-mirror/1.0")

	return d.client.Do(req)
}

func (d *Downloader) resolveURL(base *url.URL, rawRef string) (*url.URL, error) {
	ref, err := url.Parse(strings.TrimSpace(rawRef))
	if err != nil {
		return nil, err
	}

	return base.ResolveReference(ref), nil
}

func (d *Downloader) normalizeURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}

	u.Fragment = ""

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u, nil
}

func (d *Downloader) sameHost(u *url.URL) bool {
	return strings.EqualFold(u.Hostname(), d.rootHost)
}

func (d *Downloader) getVisited(rawURL string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, ok := d.visited[rawURL]
	return v, ok
}

func (d *Downloader) setVisited(rawURL, localPath string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.visited[rawURL] = localPath
}

func (d *Downloader) removeVisited(rawURL string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.visited, rawURL)
}

func (d *Downloader) localPathForPage(u *url.URL) string {
	hostDir := sanitizeHost(u.Host)

	p := u.Path
	if p == "" || p == "/" {
		p = "/index.html"
	} else if strings.HasSuffix(p, "/") {
		p = path.Join(p, "index.html")
	} else if path.Ext(p) == "" {
		p = path.Join(p, "index.html")
	}

	if u.RawQuery != "" {
		p = addQuerySuffix(p, u.RawQuery)
	}

	return filepath.Join(d.outDir, hostDir, filepath.FromSlash(strings.TrimPrefix(p, "/")))
}

func (d *Downloader) localPathForAsset(u *url.URL) string {
	hostDir := sanitizeHost(u.Host)

	p := u.Path
	if p == "" || p == "/" {
		p = "/resource"
	} else if strings.HasSuffix(p, "/") {
		p = path.Join(p, "resource")
	}

	if u.RawQuery != "" {
		p = addQuerySuffix(p, u.RawQuery)
	}

	return filepath.Join(d.outDir, hostDir, filepath.FromSlash(strings.TrimPrefix(p, "/")))
}

func isHTML(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html")
}

func writeFile(filename string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o644)
}

func relativePath(fromFile, toFile string) string {
	rel, err := filepath.Rel(filepath.Dir(fromFile), toFile)
	if err != nil {
		return toFile
	}

	return filepath.ToSlash(rel)
}

func shouldSkipURL(u *url.URL) bool {
	switch strings.ToLower(u.Scheme) {
	case "mailto", "javascript", "tel", "data":
		return true
	}
	return false
}

func sanitizeHost(host string) string {
	replacer := strings.NewReplacer(":", "_")
	return replacer.Replace(host)
}

func addQuerySuffix(p, rawQuery string) string {
	hash := shortHash(rawQuery)
	ext := path.Ext(p)

	if ext == "" {
		return p + "_" + hash
	}

	base := strings.TrimSuffix(p, ext)
	return base + "_" + hash + ext
}

func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

func getAttr(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

func setAttr(n *html.Node, key, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = value
			return
		}
	}

	n.Attr = append(n.Attr, html.Attribute{
		Key: key,
		Val: value,
	})
}
