package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Usar DATABASE_URL del backend
	dbURL := "postgres://phos_user:phos_password@localhost:5432/phosboard"
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		dbURL = envURL
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Verificar si existe El Observatodo
	fmt.Println("=== BUSCANDO EL OBSERVATODO ===")
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, url, type, fetch_strategy, interval_minutes, 
		       last_run_at, created_at, config
		FROM sources 
		WHERE url LIKE '%observatodo%' 
		   OR name LIKE '%Observatodo%'
		   OR url LIKE '%elobservatodo%'
	`)
	if err != nil {
		log.Fatal("Error querying:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var id, name, url, sourceType, fetchStrategy string
		var intervalMinutes sql.NullInt32
		var lastRunAt, createdAt sql.NullTime
		var config []byte

		if err := rows.Scan(&id, &name, &url, &sourceType, &fetchStrategy,
			&intervalMinutes, &lastRunAt, &createdAt, &config); err != nil {
			log.Fatal("Error scanning row:", err)
		}

		fmt.Printf("✅ ENCONTRADO: %s\n", name)
		fmt.Printf("   URL: %s\n", url)
		fmt.Printf("   Tipo: %s | Fetch Strategy: %s\n", sourceType, fetchStrategy)
		fmt.Printf("   Intervalo: %v minutos\n", intervalMinutes)
		fmt.Printf("   Último run: %v\n", lastRunAt)
		fmt.Printf("   Creado: %v\n", createdAt)
		fmt.Println()
	}

	if !found {
		fmt.Println("❌ NO SE ENCONTRÓ EL OBSERVATODO EN LA BASE DE DATOS")
		fmt.Println()
	}

	// 2. Verificar todos los sources
	fmt.Println("=== ESTADÍSTICAS DE SOURCES ===")
	var total, observatodo, diarioeldia int
	err = db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total_sources,
			SUM(CASE WHEN url LIKE '%observatodo%' THEN 1 ELSE 0 END) as observatodo_sources,
			SUM(CASE WHEN url LIKE '%diarioeldia%' THEN 1 ELSE 0 END) as diarioeldia_sources
		FROM sources 
		WHERE url IS NOT NULL
	`).Scan(&total, &observatodo, &diarioeldia)

	if err != nil {
		log.Fatal("Error getting stats:", err)
	}

	fmt.Printf("Total sources: %d\n", total)
	fmt.Printf("El Observatodo: %d\n", observatodo)
	fmt.Printf("Diario El Día: %d\n", diarioeldia)
	fmt.Println()

	// 3. Verificar sources que deberían ejecutarse
	fmt.Println("=== SOURCES QUE DEBERÍAN EJECUTARSE AHORA ===")
	rows, err = db.QueryContext(ctx, `
		SELECT 
			id,
			name,
			url,
			interval_minutes,
			last_run_at,
			CASE 
				WHEN last_run_at IS NULL THEN 'NUNCA EJECUTADO'
				WHEN last_run_at + (interval_minutes || ' minutes')::interval < NOW() THEN 'DEBERÍA EJECUTARSE'
				ELSE 'NO DEBERÍA EJECUTARSE TODAVÍA'
			END as status,
			last_run_at + (interval_minutes || ' minutes')::interval as next_run_time
		FROM sources 
		WHERE url IS NOT NULL 
		AND (last_run_at IS NULL OR last_run_at + (interval_minutes || ' minutes')::interval < NOW())
		ORDER BY last_run_at NULLS FIRST
		LIMIT 10
	`)
	if err != nil {
		log.Fatal("Error querying pending sources:", err)
	}
	defer rows.Close()

	pendingCount := 0
	for rows.Next() {
		pendingCount++
		var id, name, url, status string
		var intervalMinutes sql.NullInt32
		var lastRunAt, nextRunTime sql.NullTime

		if err := rows.Scan(&id, &name, &url, &intervalMinutes, &lastRunAt, &status, &nextRunTime); err != nil {
			log.Fatal("Error scanning pending row:", err)
		}

		fmt.Printf("%s: %s\n", status, name)
		fmt.Printf("   URL: %s\n", url)
		fmt.Printf("   Intervalo: %v minutos | Último: %v\n", intervalMinutes, lastRunAt)
		if nextRunTime.Valid {
			fmt.Printf("   Próximo run esperado: %v (ahora: %v)\n", nextRunTime.Time.Format("15:04:05"), time.Now().Format("15:04:05"))
		}
		fmt.Println()
	}

	if pendingCount == 0 {
		fmt.Println("✅ No hay sources pendientes de ejecución")
	} else {
		fmt.Printf("⚠️  Hay %d sources pendientes de ejecución\n", pendingCount)
	}

	// 4. Verificar configuración del dispatcher
	fmt.Println("\n=== CONFIGURACIÓN DEL DISPATCHER ===")
	fmt.Println("Intervalo por defecto: 900 segundos (15 minutos)")
	fmt.Println("El dispatcher ejecuta cada 15 minutos y busca sources donde:")
	fmt.Println("  - url IS NOT NULL")
	fmt.Println("  - last_run_at IS NULL OR last_run_at + interval_minutes < NOW()")
	fmt.Println("  - Límite: 100 sources por ciclo")
}
