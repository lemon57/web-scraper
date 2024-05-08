package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type MockProgressBar struct {
}

func (p *MockProgressBar) Add(n int) {
}

func TestResolveLink(t *testing.T) {
	testCases := map[string]struct {
		baseUrl  string
		href     string
		expected string
	}{
		"resolve regular Link": {
			"https://www.google.com",
			"/search?q=goquery",
			"https://www.google.com/search?q=goquery",
		},
		"resolve Link with extended host url": {
			"https://www.google.com/new",
			"/search?q=goquery&lang=en",
			"https://www.google.com/search?q=goquery&lang=en",
		},
		"resolve Link with absolut href": {
			"https://www.google.com/new",
			"https://www.google.com/search?q=goquery&lang=en&new=true",
			"https://www.google.com/search?q=goquery&lang=en&new=true",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := resolveLink(tc.baseUrl, tc.href)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
			assert.Equal(t, result, tc.expected)
		})
	}
}

func TestParseCssAndJsFiles(t *testing.T) {
	progressBar := &MockProgressBar{}
	mockHTML := `
			<html>
				<head>
					<link rel="stylesheet" href="styles.css">
				</head>
				<body>
					<script src="script.js"></script>
				</body>
			</html>
    	`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(mockHTML))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	response, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("Error making HTTP request: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}
	expectedResult := []string{server.URL + "/styles.css", server.URL + "/script.js"}

	assert.EqualValues(t, expectedResult, parseCssAndJsFiles(server.URL, progressBar))
}

func TestParsePageLinksNew(t *testing.T) {
	progressBar := &MockProgressBar{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
			<body>
				<a href="/page1">Page 1</a>
				<a href="/page2">Page 2</a>
				<img src="/image1.png">
				<img src="/image2.png">
			</body>
			</html>
		`))
	}))
	defer server.Close()

	visited := make(map[string]bool)
	urls := []string{}
	urls = parsePageLinks(server.URL, visited, urls, progressBar)

	expectedURLs := []string{
		server.URL,
		resolveLink(server.URL, "/page1"),
		resolveLink(server.URL, "/page2"),
		resolveLink(server.URL, "/image1.png"),
		resolveLink(server.URL, "/image2.png"),
	}
	if len(urls) != len(expectedURLs) {
		t.Fatalf("Expected %d URLs, got %d", len(expectedURLs), len(urls))
	}
	assert.EqualValues(t, expectedURLs, urls)
}

func TestScrapeWebsiteSequentially(t *testing.T) {
	progressBar := &MockProgressBar{}
	mockHTML := `
			<html>
			<body>
				<a href="/page1">Page 1</a>
				<a href="/page2">Page 2</a>
				<img src="/image1.png">
				<img src="/image2.png">
			</body>
			</html>
		`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	urls := []string{server.URL}
	scrapeWebsiteSequentially(urls, progressBar)

	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	path := dir + "/books.toscrape.com/index.html"
	fileNew, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	content := make([]byte, 165)
	_, err = fileNew.Read(content)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	assert.Equal(t, path, fileNew.Name())
	assert.Equal(t, []byte(mockHTML), content)
}
