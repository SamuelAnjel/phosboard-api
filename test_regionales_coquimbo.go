package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("=== ANÁLISIS DE PERIÓDICOS REGIONALES DE COQUIMBO ===\n")

	// Periódicos regionales de Coquimbo
	periodicos := []struct {
		nombre   string
		dominio  string
		ejemplos []string
	}{
		{
			"Diario El Día (La Serena)",
			"diarioeldia.cl",
			[]string{
				"https://www.diarioeldia.cl/noticia/nacional/2025/04/12/titulo-noticia",
				"https://www.diarioeldia.cl/deportes/2025/04/11/partido-local",
				"https://www.diarioeldia.cl/la-serena/2025/04/10/evento-municipal",
				"https://www.diarioeldia.cl/coquimbo/2025/04/09/playas-limpias",
				"https://www.diarioeldia.cl/ovalle/2025/04/08/agricultura-region",
				"https://www.diarioeldia.cl/elqui/2025/04/07/turismo-astronomico",
				"https://www.diarioeldia.cl/archivo/2024/",
				"https://www.diarioeldia.cl/contacto",
			},
		},
		{
			"El Observatodo (La Serena)",
			"elobservatodo.cl",
			[]string{
				"https://www.elobservatodo.cl/noticias/actualidad/titulo-noticia",
				"https://www.elobservatodo.cl/node/12345",
				"https://www.elobservatodo.cl/deportes/partido-regional",
				"https://www.elobservatodo.cl/cultura/festival-elqui",
				"https://www.elobservatodo.cl/region/coquimbo/noticia",
				"https://www.elobservatodo.cl/search?q=noticias",
				"https://www.elobservatodo.cl/taxonomy/term/123",
				"https://www.elobservatodo.cl/user/login",
			},
		},
		{
			"El Serenense (La Serena)",
			"elserenense.cl",
			[]string{
				"https://www.elserenense.cl/noticia/titulo-local",
				"https://www.elserenense.cl/deportes/campeonato-escolar",
				"https://www.elserenense.cl/cultura/museo-regional",
				"https://www.elserenense.cl/la-serena/proyecto-municipal",
				"https://www.elserenense.cl/archivo/2024",
				"https://www.elserenense.cl/buscar/",
			},
		},
		{
			"El Día (Ovalle)",
			"diariovallenino.cl",
			[]string{
				"https://www.diariovallenino.cl/noticia/titulo-ovalle",
				"https://www.diariovallenino.cl/limari/agricultura",
				"https://www.diariovallenino.cl/ovalle/urbano",
			},
		},
		{
			"Mi Voz La Serena",
			"laserena.mi-voz.com",
			[]string{
				"https://laserena.mi-voz.com/node/12345",
				"https://laserena.mi-voz.com/noticias/actualidad",
				"https://laserena.mi-voz.com/deportes/local",
			},
		},
	}

	// Heurísticas específicas para regionales
	heuristicas := []struct {
		nombre string
		test   func(string) bool
		puntos int
	}{
		// POSITIVAS (indican artículo)
		{"Tiene fecha (/202/)", func(url string) bool { return strings.Contains(url, "/202") }, 3},
		{"Es noticia/articulo", func(url string) bool {
			return strings.Contains(url, "noticia") || strings.Contains(url, "noticias") ||
				strings.Contains(url, "articulo") || strings.Contains(url, "article")
		}, 2},
		{"Es categoría regional", func(url string) bool {
			regionales := []string{"la-serena", "coquimbo", "ovalle", "illapel", "combarbala",
				"vicuna", "elqui", "limari", "choapa", "region", "regional", "local"}
			for _, reg := range regionales {
				if strings.Contains(url, reg) {
					return true
				}
			}
			return false
		}, 2},
		{"Es categoría general", func(url string) bool {
			generales := []string{"nacional", "internacional", "deportes", "cultura",
				"politica", "economia", "tecnologia", "ciencia", "salud", "educacion"}
			for _, gen := range generales {
				if strings.Contains(url, gen) {
					return true
				}
			}
			return false
		}, 1},
		{"Tiene ID numérico", func(url string) bool {
			// Buscar /123/ o /node/123 o ?id=123
			return strings.Contains(url, "/123") || strings.Contains(url, "?id=") ||
				strings.Contains(url, "&id=") || strings.Contains(url, "/node/")
		}, 2},

		// NEGATIVAS (indican NO artículo)
		{"Es página administrativa", func(url string) bool {
			admin := []string{"contacto", "contact", "acerca", "about", "aviso-legal",
				"terminos", "privacidad", "cookies", "login", "registro", "signup",
				"buscar", "search", "archivo", "archive", "user", "taxonomy"}
			for _, adm := range admin {
				if strings.Contains(url, adm) {
					return true
				}
			}
			return false
		}, -5},
		{"Es feed/búsqueda", func(url string) bool {
			feeds := []string{"rss", "feed", "xml", "atom", "sitemap", "mapa-del-sitio"}
			for _, feed := range feeds {
				if strings.Contains(url, feed) {
					return true
				}
			}
			return false
		}, -5},
		{"Es publicidad", func(url string) bool {
			pub := []string{"publicidad", "anuncios", "ads", "promo", "sponsored", "advertisement"}
			for _, p := range pub {
				if strings.Contains(url, p) {
					return true
				}
			}
			return false
		}, -5},
	}

	// Analizar cada periódico
	for _, periodico := range periodicos {
		fmt.Printf("📰 %s (%s)\n", periodico.nombre, periodico.dominio)
		fmt.Println(strings.Repeat("-", 50))

		articulos := 0
		noArticulos := 0

		for _, url := range periodico.ejemplos {
			puntaje := 0
			razones := []string{}

			for _, heuristica := range heuristicas {
				if heuristica.test(url) {
					puntaje += heuristica.puntos
					if heuristica.puntos > 0 {
						razones = append(razones, fmt.Sprintf("+%d: %s", heuristica.puntos, heuristica.nombre))
					} else {
						razones = append(razones, fmt.Sprintf("%d: %s", heuristica.puntos, heuristica.nombre))
					}
				}
			}

			esArticulo := puntaje >= 3
			if esArticulo {
				articulos++
			} else {
				noArticulos++
			}

			// Mostrar resultado
			icono := "✅"
			if !esArticulo {
				icono = "❌"
			}

			fmt.Printf("%s Puntaje: %2d | %s\n", icono, puntaje, url)
			if len(razones) > 0 {
				for _, razon := range razones {
					fmt.Printf("    %s\n", razon)
				}
			}
			fmt.Println()
		}

		// Resumen del periódico
		total := len(periodico.ejemplos)
		porcentaje := float64(articulos) / float64(total) * 100
		fmt.Printf("📊 Resumen: %d/%d URLs son artículos (%.0f%%)\n\n", articulos, total, porcentaje)
	}

	// Resumen general
	fmt.Println("=== PATRONES IDENTIFICADOS PARA PERIÓDICOS REGIONALES ===")
	fmt.Println("\n1. 📅 Fechas en URL: /2025/04/12/ (muy común en Diario El Día)")
	fmt.Println("2. 🏙️ Categorías regionales: /la-serena/, /coquimbo/, /ovalle/, /elqui/")
	fmt.Println("3. 📰 Términos de noticias: /noticia/, /noticias/, /articulo/")
	fmt.Println("4. 🏷️ IDs numéricos: /node/12345, /12345/, ?id=12345")
	fmt.Println("5. 🚫 Exclusiones: contacto, buscar, archivo, user, taxonomy")
	fmt.Println("6. 🏗️ Plataformas comunes: WordPress, Drupal, sistemas propios")
	fmt.Println("\n💡 Los periódicos regionales suelen tener:")
	fmt.Println("   - Estructuras más simples que los nacionales")
	fmt.Println("   - Mayor enfoque en noticias locales")
	fmt.Println("   - Menos protección anti-bots")
	fmt.Println("   - URLs más predecibles")
}
