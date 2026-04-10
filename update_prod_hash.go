package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Production database URL from Secret Manager
	dbURL := "postgresql://postgres.ohrmoiplfblbzstpgpxn:&2s9d-3cXALSPtd@aws-1-us-east-1.pooler.supabase.com:5432/postgres"
	
	ctx := context.Background()
	
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	
	// Generate correct hash for "password"
	password := "password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Failed to generate hash: %v\n", err)
		os.Exit(1)
	}
	
	hashStr := string(hash)
	fmt.Printf("Generated hash: %s\n", hashStr)
	
	// Update super admin
	result, err := conn.Exec(ctx, 
		"UPDATE users SET password_hash = $1 WHERE email = $2",
		hashStr, "admin@phosboard.cl")
	if err != nil {
		fmt.Printf("Failed to update super admin: %v\n", err)
		os.Exit(1)
	}
	
	rowsAffected := result.RowsAffected()
	fmt.Printf("Updated super admin: %d rows affected\n", rowsAffected)
	
	// Update tenant admin
	result, err = conn.Exec(ctx,
		"UPDATE users SET password_hash = $1 WHERE email = $2",
		hashStr, "tenant.admin@example.com")
	if err != nil {
		fmt.Printf("Failed to update tenant admin: %v\n", err)
		os.Exit(1)
	}
	
	rowsAffected = result.RowsAffected()
	fmt.Printf("Updated tenant admin: %d rows affected\n", rowsAffected)
	
	// Verify
	var email, hashPrefix string
	err = conn.QueryRow(ctx,
		"SELECT email, LEFT(password_hash, 30) FROM users WHERE email = $1",
		"admin@phosboard.cl").Scan(&email, &hashPrefix)
	if err != nil {
		fmt.Printf("Failed to verify: %v\n", err)
	} else {
		fmt.Printf("✅ Verified: %s -> %s...\n", email, hashPrefix)
	}
	
	fmt.Println("\n✅ Password hashes updated in production database!")
	fmt.Println("Use 'password' as password for both admin accounts")
}
