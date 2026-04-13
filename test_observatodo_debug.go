package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("=== DEBUG: ¿Por qué El Observatodo no produce URLs? ===\n")

	// URLs hipotéticas que podría encontrar El Observatodo
	exampleURLs := []string{
		// Posibles URLs reales de El Observatodo
		"https://www.elobservatodo.cl/node/12345",
		"https://www.elobservatodo.cl/node/9876",
		"https://www.elobservatodo.cl/noticias/actualidad",
		"https://www.elobservatodo.cl/deportes",
		"https://www.elobservatodo.cl/cultura",
		"https://www.elobservatodo.cl/region/coquimbo",
		"https://www.elobservatodo.cl/la-serena",
		"https://www.elobservatodo.cl/taxonomy/term/123",
		"https://www.elobservatodo.cl/user/login",
		"https://www.elobservatodo.cl/search",
		"https://www.elobservatodo.cl/rss.xml",
	}

	// Heurísticas actuales (simplificadas)
	heuristics := []struct {
		name  string
		test  func(string) bool
		score int
	}{
		// POSITIVAS
		{"Fecha /202/", func(u string) bool { return strings.Contains(u, "/202") }, 3},
		{"Noticia/articulo", func(u string) bool {
			return strings.Contains(u, "noticia") || strings.Contains(u, "articulo") ||
				strings.Contains(u, "article") || strings.Contains(u, "noticias")
		}, 2},
		{"ID numérico (/node/123)", func(u string) bool {
			// Buscar /node/ seguido de números
			if strings.Contains(u, "/node/") {
				parts := strings.Split(u, "/node/")
				if len(parts) > 1 {
					// Verificar que lo que sigue sean números
					idPart := parts[1]
					for i, ch := range idPart {
						if i >= 3 { // Solo verificar primeros 3 caracteres
							break
						}
						if ch < '0' || ch > '9' {
							return false
						}
					}
					return true
				}
			}
			return false
		}, 2},
		{"Categoría regional", func(u string) bool {
			regionales := []string{"la-serena", "coquimbo", "ovalle", "elqui", "region"}
			for _, r := range regionales {
				if strings.Contains(u, r) {
					return true
				}
			}
			return false
		}, 2},
		{"Categoría general", func(u string) bool {
			cats := []string{"deportes", "cultura", "politica", "economia", "actualidad"}
			for _, c := range cats {
				if strings.Contains(u, c) {
					return true
				}
			}
			return false
		}, 1},

		// NEGATIVAS
		{"Página administrativa", func(u string) bool {
			admin := []string{"taxonomy", "user", "search", "login", "rss", "feed"}
			for _, a := range admin {
				if strings.Contains(u, a) {
					return true
				}
			}
			return false
		}, -5},
	}

	fmt.Println("Análisis de URLs potenciales de El Observatodo:")
	fmt.Println("============================================================")

	for _, url := range exampleURLs {
		totalScore := 0
		details := []string{}

		for _, h := range heuristics {
			if h.test(url) {
				totalScore += h.score
				if h.score > 0 {
					details = append(details, fmt.Sprintf("+%d %s", h.score, h.name))
				} else {
					details = append(details, fmt.Sprintf("%d %s", h.score, h.name))
				}
			}
		}

		isUseful := totalScore >= 3
		status := "❌ NO ÚTIL"
		if isUseful {
			status = "✅ ÚTIL"
		}

		fmt.Printf("%s %s (puntaje: %d)\n", status, url, totalScore)
		if len(details) > 0 {
			fmt.Printf("  Detalles: %s\n", strings.Join(details, ", "))
		}
		fmt.Println()
	}

	// Análisis del problema
	fmt.Println("\n=== DIAGNÓSTICO DEL PROBLEMA ===")
	fmt.Println("1. El Observatodo usa Drupal con /node/[id]")
	fmt.Println("2. Nuestra heurística para /node/ requiere:")
	fmt.Println("   - /node/ seguido de al menos 3 dígitos")
	fmt.Println("   - Ejemplo: /node/123 → ✅")
	fmt.Println("   - Ejemplo: /node/abc → ❌")
	fmt.Println("   - Ejemplo: /node/12  → ❌ (solo 2 dígitos)")

	fmt.Println("\n3. Posibles problemas:")
	fmt.Println("   a) Cloudflare bloqueando el crawler")
	fmt.Println("   b) El sitio requiere JavaScript")
	fmt.Println("   c) URLs reales no coinciden con patrones esperados")
	fmt.Println("   d) El crawler no puede acceder al contenido")

	fmt.Println("\n=== SOLUCIONES SUGERIDAS ===")
	fmt.Println("1. Verificar logs de HTTP requests del crawler")
	fmt.Println("2. Ajustar heurística /node/ para aceptar 2+ dígitos")
	fmt.Println("3. Agregar User-Agents y headers específicos")
	fmt.Println("4. Probar con requests manuales para ver estructura real")
	fmt.Println("5. Verificar si hay redirects o bot protection")
}
