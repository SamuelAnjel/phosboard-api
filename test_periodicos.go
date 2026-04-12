package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"phosboard/workers/discovery/internal/discovery"
)

func main() {
	ctx := context.Background()

	// URLs de prueba para diferentes periódicos
	testURLs := []struct {
		name string
		url  string
	}{
		{"Diario El Día", "https://www.diarioeldia.cl"},
		{"Emol", "https://www.emol.com"},
		{"La Tercera", "https://www.latercera.com"},
		{"BioBioChile", "https://www.biobiochile.cl"},
		{"Cooperativa", "https://www.cooperativa.cl"},
		{"Chilevisión", "https://www.chilevision.cl"},
	}

	// Configuración base
	config := discovery.DefaultCrawlConfig()
	config.MaxDepth = 1
	config.MaxPages = 10

	for _, test := range testURLs {
		fmt.Printf("\n=== Probando %s (%s) ===\n", test.name, test.url)

		crawler := discovery.NewWebCrawler(config)

		// Timeout de 30 segundos por prueba
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		startTime := time.Now()
		results, err := crawler.Crawl(ctx, test.url)
		elapsed := time.Since(startTime)

		if err != nil {
			log.Printf("Error en %s: %v", test.name, err)
			continue
		}

		fmt.Printf("Crawl completado en %v\n", elapsed)
		fmt.Printf("URLs encontradas: %d\n", len(results))

		// Mostrar algunas URLs encontradas
		maxToShow := 5
		if len(results) < maxToShow {
			maxToShow = len(results)
		}

		for i := 0; i < maxToShow; i++ {
			fmt.Printf("  %d. %s\n", i+1, results[i].URL)
			if results[i].Title != "" {
				fmt.Printf("     Título: %s\n", results[i].Title)
			}
		}

		if len(results) > maxToShow {
			fmt.Printf("  ... y %d más\n", len(results)-maxToShow)
		}

		// Analizar tipos de URLs encontradas
		analyzeURLPatterns(results, test.name)
	}
}

func analyzeURLPatterns(results []discovery.DiscoveredURL, siteName string) {
	patterns := map[string]int{
		"con fecha (/202/)":      0,
		"noticia/noticias":       0,
		"nacional/internacional": 0,
		"deportes":               0,
		"cultura":                0,
		"politica":               0,
		"economia":               0,
		"tecnologia":             0,
		"otro":                   0,
	}

	for _, result := range results {
		url := result.URL

		found := false
		if contains(url, "/202") {
			patterns["con fecha (/202/)"]++
			found = true
		}
		if contains(url, "noticia") || contains(url, "noticias") {
			patterns["noticia/noticias"]++
			found = true
		}
		if contains(url, "nacional") || contains(url, "internacional") {
			patterns["nacional/internacional"]++
			found = true
		}
		if contains(url, "deporte") {
			patterns["deportes"]++
			found = true
		}
		if contains(url, "cultura") {
			patterns["cultura"]++
			found = true
		}
		if contains(url, "politica") {
			patterns["politica"]++
			found = true
		}
		if contains(url, "economia") || contains(url, "economía") {
			patterns["economia"]++
			found = true
		}
		if contains(url, "tecnologia") || contains(url, "tecnología") {
			patterns["tecnologia"]++
			found = true
		}

		if !found {
			patterns["otro"]++
		}
	}

	fmt.Printf("Análisis de patrones para %s:\n", siteName)
	for pattern, count := range patterns {
		if count > 0 {
			percentage := float64(count) / float64(len(results)) * 100
			fmt.Printf("  %s: %d (%.1f%%)\n", pattern, count, percentage)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			contains(s[1:], substr))))
}
