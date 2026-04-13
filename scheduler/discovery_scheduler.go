package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Source struct {
	ID              string     `json:"id"`
	URL             string     `json:"url"`
	Type            string     `json:"type"`
	IntervalMinutes int        `json:"interval_minutes"`
	LastRunAt       *time.Time `json:"last_run_at"`
}

type ProcessRequest struct {
	SourceID string `json:"source_id"`
	URL      string `json:"url"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	discoveryEndpoint := os.Getenv("DISCOVERY_ENDPOINT")
	if discoveryEndpoint == "" {
		discoveryEndpoint = "https://worker-discovery-544990213867.us-east1.run.app"
	}

	log.Printf("Starting discovery scheduler job")
	log.Printf("Database: %s", maskDBURL(dbURL))
	log.Printf("Discovery endpoint: %s", discoveryEndpoint)

	// Ejecutar un solo ciclo y terminar (para Cloud Run Jobs)
	if err := runSchedulerCycle(dbURL, discoveryEndpoint); err != nil {
		log.Fatalf("Scheduler cycle failed: %v", err)
	}

	log.Printf("Discovery scheduler job completed successfully")
}

func runSchedulerCycle(dbURL, discoveryEndpoint string) error {
	log.Printf("Starting scheduler cycle at %s", time.Now().Format("15:04:05"))

	// 1. Obtener sources pendientes
	sources, err := getPendingSources(dbURL)
	if err != nil {
		return fmt.Errorf("get pending sources: %w", err)
	}

	if len(sources) == 0 {
		log.Printf("No pending sources")
		return nil
	}

	log.Printf("Found %d pending sources", len(sources))

	// 2. Procesar cada source
	processed := 0
	for _, source := range sources {
		if err := processSource(source, discoveryEndpoint); err != nil {
			log.Printf("Failed to process source %s: %v", source.ID, err)
			continue
		}

		// 3. Actualizar last_run_at
		if err := updateLastRunAt(dbURL, source.ID); err != nil {
			log.Printf("Failed to update last_run_at for source %s: %v", source.ID, err)
			continue
		}

		processed++
		log.Printf("Processed source %s (%s)", source.ID, source.URL)
	}

	log.Printf("Scheduler cycle completed: %d/%d sources processed", processed, len(sources))
	return nil
}

func getPendingSources(dbURL string) ([]Source, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	query := `
		SELECT id, url, type, interval_minutes, last_run_at
		FROM sources 
		WHERE url IS NOT NULL 
		AND (
			-- Fuentes nunca ejecutadas (nuevas)
			last_run_at IS NULL 
			-- Fuentes fuera de su intervalo normal  
			OR last_run_at + (interval_minutes || ' minutes')::interval < NOW()
		)
		ORDER BY 
			-- Prioridad 1: Fuentes nunca ejecutadas (nuevas)
			CASE WHEN last_run_at IS NULL THEN 1 ELSE 2 END,
			-- Prioridad 2: Las más antiguas primero entre las que ya se ejecutaron
			last_run_at ASC NULLS FIRST
		LIMIT 100`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var s Source
		var lastRunAt sql.NullTime

		if err := rows.Scan(&s.ID, &s.URL, &s.Type, &s.IntervalMinutes, &lastRunAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		if lastRunAt.Valid {
			s.LastRunAt = &lastRunAt.Time
		}

		sources = append(sources, s)
	}

	return sources, rows.Err()
}

func processSource(source Source, discoveryEndpoint string) error {
	req := ProcessRequest{
		SourceID: source.ID,
		URL:      source.URL,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(discoveryEndpoint+"/process-source", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func updateLastRunAt(dbURL, sourceID string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	query := `UPDATE sources SET last_run_at = NOW() WHERE id = $1`

	_, err = db.ExecContext(ctx, query, sourceID)
	if err != nil {
		return fmt.Errorf("exec update: %w", err)
	}

	return nil
}

func maskDBURL(dbURL string) string {
	if dbURL == "" {
		return ""
	}

	parsed, err := url.Parse(dbURL)
	if err != nil {
		return "[invalid URL]"
	}

	if parsed.User != nil {
		parsed.User = url.UserPassword("***", "***")
	}

	return parsed.String()
}
