package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/jackc/pgx/v5"
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
	
	var id, email string
	var isActive bool
	err = conn.QueryRow(ctx,
		"SELECT id, email, is_active FROM users WHERE email = $1",
		"admin@phosboard.cl").Scan(&id, &email, &isActive)
	if err != nil {
		fmt.Printf("Failed to query user: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("User: %s (%s)\n", email, id)
	fmt.Printf("Is active: %v\n", isActive)
	
	if !isActive {
		fmt.Println("❌ User is NOT active! This could be the problem.")
		// Try to activate
		_, err = conn.Exec(ctx, "UPDATE users SET is_active = TRUE WHERE id = $1", id)
		if err != nil {
			fmt.Printf("Failed to activate user: %v\n", err)
		} else {
			fmt.Println("✅ User activated")
		}
	} else {
		fmt.Println("✅ User is active")
	}
}
