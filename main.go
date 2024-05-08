package main

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const startURL = "https://books.toscrape.com/"

var countBar *progressbar.ProgressBar

func main() {
	visited := make(map[string]bool)
	urls := make([]string, 0)
	countBar = progressbar.Default(-1, "Counting links to parse a website "+startURL)
	urls = parseCssAndJsFiles(startURL)
	urls, _ = parsePageLinks(startURL, visited, urls)
	scrapeWebsite(urls)
}

func parsePageLinks(u string, visited map[string]bool, urls []string) (page []string, err error) {
	elementMatcher := map[string]string{
		"a":   "href",
		"img": "src",
	}
	if visited[u] {
		return urls, nil
	}
	visited[u] = true
	urls = append(urls, u)
	countBar.Add(1)

	// Get the HTML
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println(err)
		return urls, err
	}
	defer resp.Body.Close()

	// Create a goquery document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
		return urls, err
	}

	// Find and visit all pages and images
	for selector, attr := range elementMatcher {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr(attr)
			if exists {
				link := resolveLink(u, href)
				urls, _ = parsePageLinks(link, visited, urls)
			}
		})
	}

	return urls, nil
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
				countBar.Add(1)
			}
		})
	}

	return urls
}

func scrapeWebsite(urls []string) {
	count := len(urls)
	fmt.Println("\nBeginning scraping a website...")
	start := time.Now()
	progressBar := progressbar.Default(int64(count), "Scraping a website "+startURL)
	for _, link := range urls {
		resp, err := http.Get(link)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		savePage(link, resp.Body)
		progressBar.Add(1)
	}
	duration := time.Since(start)
	fmt.Println("Scraping time: ", duration.Minutes(), " min")
	fmt.Println("Total downloaded files: ", count)
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
