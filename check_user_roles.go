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
	
	// Check user
	var userID, email string
	err = conn.QueryRow(ctx,
		"SELECT id, email FROM users WHERE email = $1",
		"admin@phosboard.cl").Scan(&userID, &email)
	if err != nil {
		fmt.Printf("Failed to find user: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("User: %s (%s)\n", email, userID)
	
	// Check user roles
	rows, err := conn.Query(ctx, `
		SELECT r.name, ur.tenant_id
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		fmt.Printf("Failed to query roles: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()
	
	var roleCount int
	for rows.Next() {
		var roleName string
		var tenantID *string
		err = rows.Scan(&roleName, &tenantID)
		if err != nil {
			fmt.Printf("Failed to scan role: %v\n", err)
			continue
		}
		
		tenantStr := "NULL"
		if tenantID != nil {
			tenantStr = *tenantID
		}
		fmt.Printf("  Role: %s, Tenant: %s\n", roleName, tenantStr)
		roleCount++
	}
	
	if roleCount == 0 {
		fmt.Println("❌ User has NO roles assigned!")
	} else {
		fmt.Printf("✅ User has %d role(s)\n", roleCount)
	}
}
