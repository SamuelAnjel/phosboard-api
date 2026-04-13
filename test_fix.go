package main

import (
	"fmt"
	"net/url"
	"strings"
)

func resolveURL(base *url.URL, relative string) string {
	relative = strings.TrimSpace(relative)
	
	if strings.Contains(relative, "https://") {
		count := strings.Count(relative, "https://")
		if count > 1 {
			firstIdx := strings.Index(relative, "https://")
			secondIdx := strings.Index(relative[firstIdx+1:], "https://")
			if secondIdx > 0 {
				realURL := relative[firstIdx+1+secondIdx:]
				relative = realURL
			}
		}
	}
	
	if strings.Contains(relative, "http://") {
		count := strings.Count(relative, "http://")
		if count > 1 {
			firstIdx := strings.Index(relative, "http://")
			secondIdx := strings.Index(relative[firstIdx+1:], "http://")
			if secondIdx > 0 {
				realURL := relative[firstIdx+1+secondIdx:]
				relative = realURL
			}
		}
	}
	
	if strings.HasPrefix(relative, "http://") || strings.HasPrefix(relative, "https://") {
		return relative
	}
	
	if strings.HasPrefix(relative, "//") {
		return base.Scheme + ":" + relative
	}
	
	relURL, err := url.Parse(relative)
	if err != nil {
		relURL = &url.URL{Path: relative}
	}
	
	resolved := base.ResolveReference(relURL)
	return resolved.String()
}

func main() {
	base, _ := url.Parse("https://www.diarioeldia.cl/")
	
	tests := []struct{
		input string
		expected string
	}{
		{"https://www.diarioeldia.cl/https://papeldigital.eldia.la/2025/04/13", "https://papeldigital.eldia.la/2025/04/13"},
		{"/noticia/politica", "https://www.diarioeldia.cl/noticia/politica"},
		{"https://papeldigital.eldia.la/", "https://papeldigital.eldia.la/"},
		{"noticia/local", "https://www.diarioeldia.cl/noticia/local"},
		{"//cdn.example.com/image.jpg", "https://cdn.example.com/image.jpg"},
	}
	
	for _, test := range tests {
		result := resolveURL(base, test.input)
		status := "✓"
		if result != test.expected {
			status = "✗"
		}
		fmt.Printf("%s %s -> %s (expected: %s)\n", status, test.input, result, test.expected)
	}
}
