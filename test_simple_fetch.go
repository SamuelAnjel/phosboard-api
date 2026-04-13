package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	url := "https://www.miradiols.cl/2026/04/01/tenemos-un-cementerio-de-moviles-y-equipos-sin-funcionar-afusam-critica-dichos-de-alcaldesa-por-millones-sobrantes-en-municipio-de-la-serena/"

	fmt.Println("Testing with Accept-Encoding: gzip, deflate (no brotli)...")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate") // NO brotli
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Content-Encoding: %s\n", resp.Header.Get("Content-Encoding"))
	fmt.Printf("Body size: %d bytes\n", len(body))

	// Handle decompression
	contentEncoding := resp.Header.Get("Content-Encoding")
	finalBody := body

	if strings.Contains(strings.ToLower(contentEncoding), "gzip") {
		fmt.Println("Decompressing gzip...")
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			fmt.Printf("Failed to create gzip reader: %v\n", err)
		} else {
			decompressed, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				fmt.Printf("Failed to decompress: %v\n", err)
			} else {
				finalBody = decompressed
				fmt.Printf("Decompressed to: %d bytes\n", len(finalBody))
			}
		}
	}

	// Save to file
	filename := "/tmp/test_no_brotli.html"
	os.WriteFile(filename, finalBody, 0644)
	fmt.Printf("Saved to: %s\n", filename)

	// Check if it's HTML
	finalStr := string(finalBody)
	isHTML := strings.Contains(finalStr, "<!DOCTYPE") ||
		strings.Contains(finalStr, "<html") ||
		strings.Contains(finalStr, "<body")

	fmt.Printf("Is HTML: %v\n", isHTML)

	if isHTML {
		// Show first 200 chars
		if len(finalStr) > 200 {
			fmt.Println("\nFirst 200 characters:")
			fmt.Println(finalStr[:200])
		}

		// Check for Cloudflare
		if strings.Contains(strings.ToLower(finalStr), "cloudflare") ||
			strings.Contains(strings.ToLower(finalStr), "challenge") ||
			strings.Contains(strings.ToLower(finalStr), "captcha") {
			fmt.Println("\nWARNING: Cloudflare challenge page detected!")
		} else {
			fmt.Println("\nLooks like real HTML content (not Cloudflare challenge)")
		}
	} else {
		// Show hex dump
		fmt.Println("\nFirst 50 bytes (hex):")
		for i := 0; i < min(50, len(finalBody)); i += 16 {
			end := min(i+16, len(finalBody))
			for j := i; j < end; j++ {
				fmt.Printf("%02x ", finalBody[j])
			}
			fmt.Println()
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
