package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

func main() {
	url := "https://www.miradiols.cl/2026/04/01/tenemos-un-cementerio-de-moviles-y-equipos-sin-funcionar-afusam-critica-dichos-de-alcaldesa-por-millones-sobrantes-en-municipio-de-la-serena/"

	fmt.Println("Testing scraper extraction on:", url)

	// Create HTTP client with browser-like headers
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Set browser-like headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching URL: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Read body with proper encoding detection
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return
	}

	fmt.Printf("Body size: %d bytes\n", len(bodyBytes))

	// Save raw HTML
	rawFile := "/tmp/raw_test.html"
	if err := os.WriteFile(rawFile, bodyBytes, 0644); err != nil {
		fmt.Printf("Error saving raw HTML: %v\n", err)
	} else {
		fmt.Printf("Raw HTML saved to: %s\n", rawFile)
	}

	// Convert to UTF-8
	utf8Reader, err := charset.NewReader(bytes.NewReader(bodyBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		fmt.Printf("Error converting to UTF-8: %v\n", err)
		return
	}

	utf8Bytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		fmt.Printf("Error reading UTF-8 content: %v\n", err)
		return
	}

	fmt.Printf("UTF-8 size: %d bytes\n", len(utf8Bytes))

	// Save UTF-8 HTML
	utf8File := "/tmp/utf8_test.html"
	if err := os.WriteFile(utf8File, utf8Bytes, 0644); err != nil {
		fmt.Printf("Error saving UTF-8 HTML: %v\n", err)
	} else {
		fmt.Printf("UTF-8 HTML saved to: %s\n", utf8File)
	}

	// Extract clean HTML (remove scripts, styles, etc.)
	cleanHTML := extractCleanHTML(string(utf8Bytes))

	fmt.Printf("Clean HTML size: %d bytes\n", len(cleanHTML))

	// Save clean HTML
	cleanFile := "/tmp/clean_test.html"
	if err := os.WriteFile(cleanFile, []byte(cleanHTML), 0644); err != nil {
		fmt.Printf("Error saving clean HTML: %v\n", err)
	} else {
		fmt.Printf("Clean HTML saved to: %s\n", cleanFile)
	}

	// Extract plain text
	plainText := extractPlainText(cleanHTML)

	fmt.Printf("Plain text size: %d bytes\n", len(plainText))

	// Save plain text
	textFile := "/tmp/text_test.txt"
	if err := os.WriteFile(textFile, []byte(plainText), 0644); err != nil {
		fmt.Printf("Error saving plain text: %v\n", err)
	} else {
		fmt.Printf("Plain text saved to: %s\n", textFile)
	}

	// Show first 500 chars of text
	if len(plainText) > 500 {
		fmt.Println("\nFirst 500 chars of text:")
		fmt.Println(plainText[:500])
	} else {
		fmt.Println("\nFull text:")
		fmt.Println(plainText)
	}

	// Check for encoding issues
	fmt.Println("\nChecking for encoding issues...")
	if strings.Contains(string(utf8Bytes), "Ã") || strings.Contains(string(utf8Bytes), "Â") {
		fmt.Println("WARNING: Found double-encoded characters (Ã, Â)")
	}

	// Check for binary/corrupt data
	asciiCount := 0
	for _, b := range plainText {
		if b >= 32 && b <= 126 || b == '\n' || b == '\t' || b == '\r' {
			asciiCount++
		}
	}

	asciiPercent := float64(asciiCount) / float64(len(plainText)) * 100
	fmt.Printf("ASCII characters in text: %.1f%%\n", asciiPercent)

	if asciiPercent < 80 {
		fmt.Println("WARNING: Low ASCII percentage - possible binary/corrupt data")
	}
}

func extractCleanHTML(html string) string {
	// Remove script tags
	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	html = reScript.ReplaceAllString(html, "")

	// Remove style tags
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = reStyle.ReplaceAllString(html, "")

	// Remove noscript tags
	reNoscript := regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`)
	html = reNoscript.ReplaceAllString(html, "")

	// Remove comments
	reComment := regexp.MustCompile(`(?is)<!--.*?-->`)
	html = reComment.ReplaceAllString(html, "")

	return html
}

func extractPlainText(html string) string {
	// Simple HTML to text conversion
	reTags := regexp.MustCompile(`(?is)<[^>]+>`)
	text := reTags.ReplaceAllString(html, " ")

	// Collapse multiple spaces
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	// Trim
	text = strings.TrimSpace(text)

	return text
}
