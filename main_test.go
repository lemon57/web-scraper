package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
