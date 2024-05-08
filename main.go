package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	startURL := "https://books.toscrape.com/"

	visited := make(map[string]bool)
	urls := make([]string, 0)

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

		// Create a goquery document
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Find and visit all links
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				link := resolveLink(u, href)
				visit(link)
			}
		})
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("src")
			if exists {
				link := resolveLink(u, href)
				visit(link)
			}
		})
	}
	visit(startURL)

	for i, u := range urls {
		fmt.Println(i, u)
	}
	fmt.Println("Visited ", len(urls), len(visited), "pages")
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
