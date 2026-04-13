package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func removeScriptsAndStylesSimple(html []byte) string {
	htmlStr := string(html)

	// Remove script tags (including content)
	reScript := regexp.MustCompile(`<script\b[^>]*>[\s\S]*?</script>`)
	htmlStr = reScript.ReplaceAllString(htmlStr, "")

	// Remove style tags (including content)
	reStyle := regexp.MustCompile(`<style\b[^>]*>[\s\S]*?</style>`)
	htmlStr = reStyle.ReplaceAllString(htmlStr, "")

	// Remove iframe tags
	reIframe := regexp.MustCompile(`<iframe\b[^>]*>[\s\S]*?</iframe>`)
	htmlStr = reIframe.ReplaceAllString(htmlStr, "")

	// Remove noscript tags
	reNoscript := regexp.MustCompile(`<noscript\b[^>]*>[\s\S]*?</noscript>`)
	htmlStr = reNoscript.ReplaceAllString(htmlStr, "")

	// Remove common ad/analytics patterns
	adPatterns := []string{
		`<!--\s*Ad\s*-->`,
		`<!--\s*Google\s*Analytics\s*-->`,
		`<!--\s*Facebook\s*Pixel\s*-->`,
		`<div[^>]*class="[^"]*\bad\b[^"]*"[^>]*>[\s\S]*?</div>`,
		`<div[^>]*id="[^"]*\bad\b[^"]*"[^>]*>[\s\S]*?</div>`,
	}

	for _, pattern := range adPatterns {
		re := regexp.MustCompile(pattern)
		htmlStr = re.ReplaceAllString(htmlStr, "")
	}

	// Clean up extra whitespace
	reMultiSpace := regexp.MustCompile(`\s+`)
	htmlStr = reMultiSpace.ReplaceAllString(htmlStr, " ")

	reMultiNewline := regexp.MustCompile(`\n\s*\n\s*\n+`)
	htmlStr = reMultiNewline.ReplaceAllString(htmlStr, "\n\n")

	return strings.TrimSpace(htmlStr)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_regex.go <html_file>")
		os.Exit(1)
	}

	htmlFile := os.Args[1]
	html, err := ioutil.ReadFile(htmlFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original HTML: %d bytes\n", len(html))

	// Test the regex cleaning
	cleaned := removeScriptsAndStylesSimple(html)

	fmt.Printf("Cleaned HTML: %d bytes\n", len(cleaned))
	fmt.Printf("Reduction: %.1f%%\n", 100-float64(len(cleaned))*100/float64(len(html)))

	// Check key markers
	fmt.Println("\n=== MARKER CHECK ===")
	markers := []struct {
		name string
		text string
	}{
		{"DOCTYPE", "<!DOCTYPE"},
		{"HTML tag", "<html"},
		{"Article title", "cementerio de móviles"},
		{"Article content", "Tenemos un cementerio de móviles en los centros"},
		{"Related article", "Deportes La Serena recibirá"},
		{"Closing html", "</html>"},
	}

	for _, marker := range markers {
		if strings.Contains(cleaned, marker.text) {
			fmt.Printf("✓ Found: %s\n", marker.name)
		} else {
			fmt.Printf("✗ Missing: %s\n", marker.name)
		}
	}

	// Show sample
	fmt.Println("\n=== FIRST 500 CHARS ===")
	if len(cleaned) > 500 {
		fmt.Println(cleaned[:500])
	} else {
		fmt.Println(cleaned)
	}

	fmt.Println("\n=== LAST 500 CHARS ===")
	if len(cleaned) > 500 {
		fmt.Println(cleaned[len(cleaned)-500:])
	} else {
		fmt.Println(cleaned)
	}

	// Check if it's the problematic fragment
	fmt.Println("\n=== PROBLEMATIC FRAGMENT CHECK ===")
	if strings.Contains(cleaned, `<div class="jeg_thumb">`) {
		// Find where it starts
		idx := strings.Index(cleaned, `<div class="jeg_thumb">`)
		// Show 1000 chars from there
		end := idx + 1000
		if end > len(cleaned) {
			end = len(cleaned)
		}
		fragment := cleaned[idx:end]

		// Check if this is the ONLY content
		lines := strings.Split(cleaned, "\n")
		fmt.Printf("Total lines: %d\n", len(lines))
		fmt.Printf("Fragment at position: %d\n", idx)
		fmt.Printf("Fragment preview: %.200s...\n", fragment)

		// Count occurrences of key elements
		articleCount := strings.Count(cleaned, "<article")
		divCount := strings.Count(cleaned, "<div")
		fmt.Printf("Article tags: %d\n", articleCount)
		fmt.Printf("Div tags: %d\n", divCount)
	}
}
