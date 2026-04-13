package main

import (
	"fmt"
	"net/url"
	"strings"
)

func resolveURL(base *url.URL, relative string) string {
	relative = strings.TrimSpace(relative)
	
	if strings.HasPrefix(relative, "http://") || strings.HasPrefix(relative, "https://") {
		parsed, err := url.Parse(relative)
		if err == nil {
			if strings.Contains(parsed.Path, "http://") || strings.Contains(parsed.Path, "https://") {
				path := parsed.Path
				var protocolIdx int
				if idx := strings.Index(path, "http://"); idx > 0 {
					protocolIdx = idx
				} else if idx := strings.Index(path, "https://"); idx > 0 {
					protocolIdx = idx
				}
				
				if protocolIdx > 0 {
					realURL := path[protocolIdx:]
					if parsedRealURL, err := url.Parse(realURL); err == nil && parsedRealURL.Scheme != "" {
						return realURL
					}
				}
			}
		}
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
		description string
	}{
		{"https://www.diarioeldia.cl/https://papeldigital.eldia.la/2025/04/13", "https://papeldigital.eldia.la/2025/04/13", "URL concatenada"},
		{"/noticia/http://example.com", "https://www.diarioeldia.cl/noticia/http://example.com", "Path con http:// en medio"},
		{"https://papeldigital.eldia.la/", "https://papeldigital.eldia.la/", "URL absoluta normal"},
		{"/noticia/politica", "https://www.diarioeldia.cl/noticia/politica", "Path relativo"},
		{"noticias", "https://www.diarioeldia.cl/noticias", "Path sin slash"},
	}
	
	for _, test := range tests {
		result := resolveURL(base, test.input)
		status := "✓"
		if result != test.expected {
			status = "✗"
		}
		fmt.Printf("%s %s\n", status, test.description)
		fmt.Printf("  Input: %s\n", test.input)
		fmt.Printf("  Output: %s\n", result)
		fmt.Printf("  Expected: %s\n\n", test.expected)
	}
}
