package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := "postgresql://postgres.ohrmoiplfblbzstpgpxn:&2s9d-3cXALSPtd@aws-1-us-east-1.pooler.supabase.com:5432/postgres"
	
	ctx := context.Background()
	
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	
	var email, hash string
	err = conn.QueryRow(ctx,
		"SELECT email, password_hash FROM users WHERE email = $1",
		"admin@phosboard.cl").Scan(&email, &hash)
	if err != nil {
		fmt.Printf("Failed to query: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("User: %s\n", email)
	fmt.Printf("Hash: %s\n", hash)
	fmt.Printf("Hash length: %d\n", len(hash))
	
	// Try to verify with "password"
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("password"))
	if err != nil {
		fmt.Printf("❌ Hash verification failed: %v\n", err)
		
		// Try with common variations
		for _, pw := range []string{"Password", "PASSWORD", "admin123", "Admin123"} {
			err2 := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
			if err2 == nil {
				fmt.Printf("✅ Hash matches: '%s'\n", pw)
				return
			}
		}
	} else {
		fmt.Println("✅ Hash matches 'password'")
	}
}
