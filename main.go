package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	startURL := "https://books.toscrape.com/"
	start := time.Now()
	visited := make(map[string]bool)
	urls := make([]string, 0)
	elementMatcher := map[string]string{
		"a":   "href",
		"img": "src",
	}
	urls = parseCssAndJsFiles(startURL)

	var visit func(string)
	visit = func(u string) {
		if visited[u] {
			return
		}
		visited[u] = true
		urls = append(urls, u)

		// Get the HTML
		resp, err := http.Get(u)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		file := savePage(u, resp.Body)

		// Create a goquery document
		doc, err := goquery.NewDocumentFromReader(file)

		if err != nil {
			fmt.Println(err)
			return
		}

		// Find and visit all pages and images
		for selector, attr := range elementMatcher {
			doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				href, exists := s.Attr(attr)
				if exists {
					link := resolveLink(u, href)
					visit(link)
				}
			})
		}
	}
	visit(startURL)

	for i, u := range urls {
		fmt.Println(i, u)
	}
	duration := time.Since(start)
	fmt.Println("Scraping time: ", duration.Seconds(), " sec")
	fmt.Println("Visited ", len(urls), len(visited), "pages")
}

func parseCssAndJsFiles(u string) []string {
	var urls []string
	elementMatcher := map[string]string{
		"link":   "href",
		"script": "src",
	}

	resp, err := http.Get(u)
	if err != nil {
		fmt.Println(err)
		return urls
	}
	defer resp.Body.Close()

	// Create a goquery document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
		return urls
	}

	// Parse all CSS and script files
	for selector, attr := range elementMatcher {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			css, exists := s.Attr(attr)
			if exists {
				link := resolveLink(u, css)
				urls = append(urls, link)
			}
		})
	}

	return urls
}

func resolveLink(baseURL, href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	return base.ResolveReference(u).String()
}

func savePage(u string, body io.Reader) *os.File {
	parsedURL, _ := url.Parse(u)
	path := parsedURL.Path
	if path == "" || strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	path = filepath.Join("books.toscrape.com", path)
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fileNew, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return fileNew
}
