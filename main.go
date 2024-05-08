package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	startURL := "https://books.toscrape.com/"
	var count int

	// Get the HTML
	resp, err := http.Get(startURL)
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
			link := resolveLink(startURL, href)
			count++
			fmt.Println(count, link)
		}
	})
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
