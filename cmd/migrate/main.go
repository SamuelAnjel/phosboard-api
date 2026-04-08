package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env: %v\n", err)
	}

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("DATABASE_URL not set")
		os.Exit(1)
	}

	fmt.Printf("Connecting to: %s\n", dbURL[:30]+"...")

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			fmt.Printf("Error closing connection: %v\n", err)
		}
	}()

	files, err := filepath.Glob("data/migrations/*.sql")
	if err != nil {
		fmt.Printf("Failed to find migrations: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		files, err = filepath.Glob("/app/data/migrations/*.sql")
		if err != nil {
			fmt.Printf("Failed to find migrations in /app: %v\n", err)
			os.Exit(1)
		}
	}
	if err != nil {
		fmt.Printf("Failed to find migrations: %v\n", err)
		os.Exit(1)
	}

	sort.Strings(files)

	fmt.Printf("Found %d migration files\n", len(files))

	for _, file := range files {
		fmt.Printf("Running %s...\n", filepath.Base(file))

		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", file, err)
			continue
		}

		_, err = conn.Exec(ctx, string(content))
		if err != nil {
			fmt.Printf("Failed to run %s: %v\n", file, err)
			continue
		}

		fmt.Printf("✅ %s completed\n", filepath.Base(file))
	}

	fmt.Println("All migrations completed!")
}
