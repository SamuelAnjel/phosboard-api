package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()
	
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("DATABASE_URL not set")
		os.Exit(1)
	}
	
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	
	password := "password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Failed to generate hash: %v\n", err)
		os.Exit(1)
	}
	
	// Update super admin
	_, err = conn.Exec(ctx, 
		"UPDATE users SET password_hash = $1 WHERE email = $2",
		string(hash), "admin@phosboard.cl")
	if err != nil {
		fmt.Printf("Failed to update super admin: %v\n", err)
		os.Exit(1)
	}
	
	// Update tenant admin
	_, err = conn.Exec(ctx,
		"UPDATE users SET password_hash = $1 WHERE email = $2",
		string(hash), "tenant.admin@example.com")
	if err != nil {
		fmt.Printf("Failed to update tenant admin: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Updated passwords for both admin users to: '%s'\n", password)
	fmt.Printf("Hash: %s\n", string(hash))
}
