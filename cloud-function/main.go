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

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Source struct {
	ID              string `json:"id"`
	URL             string `json:"url"`
	Type            string `json:"type"`
	IntervalMinutes int    `json:"interval_minutes"`
}

type ProcessRequest struct {
	SourceID string `json:"source_id"`
	URL      string `json:"url"`
}

func DiscoveryScheduler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		http.Error(w, "DATABASE_URL not configured", http.StatusInternalServerError)
		return
	}

	discoveryEndpoint := os.Getenv("DISCOVERY_ENDPOINT")
	if discoveryEndpoint == "" {
		discoveryEndpoint = "https://worker-discovery-544990213867.us-east1.run.app"
	}

	log.Printf("Starting discovery scheduler cycle")

	// 1. Obtener sources pendientes
	sources, err := getPendingSources(dbURL)
	if err != nil {
		log.Printf("Error getting pending sources: %v", err)
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	if len(sources) == 0 {
		log.Printf("No pending sources")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"no pending sources"}`))
		return
	}

	log.Printf("Found %d pending sources", len(sources))

	// 2. Procesar cada source
	processed := 0
	errors := []string{}
	for _, source := range sources {
		if err := processSource(source, discoveryEndpoint); err != nil {
			errorMsg := fmt.Sprintf("source %s: %v", source.ID, err)
			log.Printf("Error: %s", errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		// 3. Actualizar last_run_at
		if err := updateLastRunAt(dbURL, source.ID); err != nil {
			errorMsg := fmt.Sprintf("update last_run_at for %s: %v", source.ID, err)
			log.Printf("Error: %s", errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		processed++
		log.Printf("Processed source %s (%s)", source.ID, source.URL)
	}

	response := map[string]interface{}{
		"status":   "completed",
		"total":    len(sources),
		"processed": processed,
		"errors":   errors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Scheduler cycle completed: %d/%d sources processed", processed, len(sources))
}

func getPendingSources(dbURL string) ([]Source, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	query := `
		SELECT id, url, type, interval_minutes
		FROM sources 
		WHERE url IS NOT NULL 
		AND (
			last_run_at IS NULL 
			OR last_run_at + (interval_minutes || ' minutes')::interval < NOW()
		)
		ORDER BY 
			CASE WHEN last_run_at IS NULL THEN 1 ELSE 2 END,
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
		if err := rows.Scan(&s.ID, &s.URL, &s.Type, &s.IntervalMinutes); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
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

func main() {
	http.HandleFunc("/", DiscoveryScheduler)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting discovery scheduler function on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
