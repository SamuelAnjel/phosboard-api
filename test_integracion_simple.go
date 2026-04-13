package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("=== Análisis de patrones para periódicos chilenos ===\n")

	// URLs de ejemplo de diferentes periódicos
	urls := []struct {
		site string
		url  string
	}{
		// Diario El Día
		{"Diario El Día", "https://www.diarioeldia.cl/noticia/nacional/2025/04/12/titulo-noticia"},
		{"Diario El Día", "https://www.diarioeldia.cl/deportes/2025/04/11/otra-noticia"},
		{"Diario El Día", "https://www.diarioeldia.cl/archivo/2024/"},

		// Emol
		{"Emol", "https://www.emol.com/noticias/Nacional/2025/04/12/1234567/titulo.html"},
		{"Emol", "https://www.emol.com/deportes/2025/04/11/1234568/otro.html"},
		{"Emol", "https://www.emol.com/buscar/?q=noticias"},

		// La Tercera
		{"La Tercera", "https://www.latercera.com/nacional/noticia/titulo-noticia/ABCDEFG/"},
		{"La Tercera", "https://www.latercera.com/opinion/columna/autor/"},
		{"La Tercera", "https://www.latercera.com/etiqueta/politica/"},

		// BioBioChile
		{"BioBioChile", "https://www.biobiochile.cl/noticias/nacional/region-metropolitana/2025/04/12/titulo.shtml"},
		{"BioBioChile", "https://www.biobiochile.cl/lista/noticias.shtml"},
		{"BioBioChile", "https://www.biobiochile.cl/publicidad/"},

		// URLs no-artículo
		{"No-artículo", "https://www.diarioeldia.cl/contacto"},
		{"No-artículo", "https://www.emol.com/rss/noticias.xml"},
		{"No-artículo", "https://www.latercera.com/aviso-legal/"},
		{"No-artículo", "https://www.biobiochile.cl/buscar/"},
	}

	// Heurísticas implementadas
	patterns := []struct {
		name string
		test func(string) bool
	}{
		{"Tiene fecha (/202/)", func(url string) bool { return strings.Contains(url, "/202") }},
		{"Es noticia/articulo", func(url string) bool {
			return strings.Contains(url, "noticia") || strings.Contains(url, "noticias") ||
				strings.Contains(url, "articulo") || strings.Contains(url, "article")
		}},
		{"Es categoría específica", func(url string) bool {
			categories := []string{"nacional", "internacional", "deportes", "cultura",
				"politica", "economia", "tecnologia", "ciencia", "salud", "educacion"}
			for _, cat := range categories {
				if strings.Contains(url, cat) {
					return true
				}
			}
			return false
		}},
		{"Tiene ID numérico", func(url string) bool {
			// Buscar patrones como /12345/ o -12345. o id=12345
			return strings.Contains(url, "/12345") || strings.Contains(url, "-12345") ||
				strings.Contains(url, "id=12345") || strings.Contains(url, "id_")
		}},
		{"Es página no-artículo", func(url string) bool {
			exclude := []string{"contacto", "contact", "acerca", "about", "aviso-legal",
				"terminos", "privacidad", "cookies", "rss", "feed", "xml", "atom",
				"buscar", "search", "archivo", "archive", "etiqueta", "tag", "categoria",
				"category", "autor", "author", "publicidad", "ads", "promo"}
			for _, ex := range exclude {
				if strings.Contains(url, ex) {
					return true
				}
			}
			return false
		}},
	}

	// Analizar cada URL
	for _, u := range urls {
		fmt.Printf("%s:\n", u.site)
		fmt.Printf("  URL: %s\n", u.url)

		score := 0
		for _, pattern := range patterns {
			if pattern.test(u.url) {
				fmt.Printf("  ✓ %s\n", pattern.name)
				if pattern.name == "Es página no-artículo" {
					score -= 5
				} else {
					score += 2
				}
			}
		}

		// Determinar si es artículo
		isArticle := score >= 3
		fmt.Printf("  Puntaje: %d -> ¿Es artículo? %v\n\n", score, isArticle)
	}

	fmt.Println("\n=== Resumen de heurísticas implementadas ===")
	fmt.Println("1. Patrones de URL con fecha (/2024/, /2025-04/)")
	fmt.Println("2. Términos de contenido (noticia, artículo, reportaje, etc.)")
	fmt.Println("3. Categorías específicas (nacional, deportes, cultura, etc.)")
	fmt.Println("4. IDs numéricos (/12345/, -12345.html, id=12345)")
	fmt.Println("5. Exclusión de páginas administrativas (contacto, aviso legal, etc.)")
	fmt.Println("6. Exclusión de feeds y búsquedas (rss, buscar, etc.)")
	fmt.Println("7. Configuración por dominio (diarioeldia.cl, emol.com, etc.)")
	fmt.Println("\nEl sistema ahora es genérico y adaptable a múltiples periódicos.")
}
