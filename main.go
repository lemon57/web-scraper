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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const startURL = "https://books.toscrape.com/"

var progressBar *progressbar.ProgressBar

func main() {
	visited := make(map[string]bool)
	urls := make([]string, 0)
	progressBar = progressbar.Default(-1, "Counting links to parse a website "+startURL)
	urls = parseCssAndJsFiles(startURL)
	urls, err := parsePageLinks(startURL, visited, urls)
	if err != nil {
		fmt.Println(err)
		return
	}
	scrapeWebsiteSequentially(urls)
	scrapeWebsiteWithMultiThreading(urls)
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
	progressBar.Add(1)
	urls = append(urls, u)

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
				progressBar.Add(1)
			}
		})
	}

	return urls
}

func scrapeWebsiteSequentially(urls []string) {
	count := len(urls)
	fmt.Println("\nTotal links to scrape: ", count)
	fmt.Println("Beginning scraping a website...")
	start := time.Now()
	progressBar = progressbar.Default(int64(count), "Scraping a website "+startURL)
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

func scrapeWebsiteWithMultiThreading(urls []string) {
	count := len(urls)
	progressBar = progressbar.Default(int64(count), "Scraping a website with multi threading "+startURL)
	start := time.Now()
	urlChan := make(chan string)
	var wg sync.WaitGroup

	// Start multiple worker goroutines
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range urlChan {
				resp, err := http.Get(link)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
				savePage(link, resp.Body)
				progressBar.Add(1)
			}
		}()
	}

	// Feed the URLs to the channel
	for _, link := range urls {
		urlChan <- link
	}
	close(urlChan)
	wg.Wait()

	duration := time.Since(start)
	fmt.Println("Scraping time: ", duration.Seconds(), " sec")
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

func savePage(u string, body io.Reader) {
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
		return
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		fmt.Println(err)
		return
	}
}
