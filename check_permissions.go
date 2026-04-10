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
	
	// Check tenant-admin role permissions
	fmt.Println("Checking permissions for 'tenant-admin' role...")
	
	rows, err := conn.Query(ctx, `
		SELECT p.endpoint, p.method, p.description
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		JOIN roles r ON rp.role_id = r.id
		WHERE r.name = 'tenant-admin'
		ORDER BY p.endpoint, p.method
	`)
	if err != nil {
		fmt.Printf("Failed to query permissions: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()
	
	var count int
	for rows.Next() {
		var endpoint, method, description string
		err = rows.Scan(&endpoint, &method, &description)
		if err != nil {
			fmt.Printf("Failed to scan: %v\n", err)
			continue
		}
		
		fmt.Printf("  %-6s %-30s %s\n", method, endpoint, description)
		count++
	}
	
	fmt.Printf("\nTotal permissions for tenant-admin: %d\n", count)
	
	// Check what permission would be needed for POST /api/v1/documents/track
	fmt.Println("\nChecking what permission is needed for POST /api/v1/documents/track...")
	
	var hasPermission bool
	err = conn.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM permissions
			WHERE endpoint = '/api/v1/documents/track'
			AND method = 'POST'
		)
	`).Scan(&hasPermission)
	if err != nil {
		fmt.Printf("Failed to check: %v\n", err)
	} else if hasPermission {
		fmt.Println("✅ Permission exists in database")
	} else {
		fmt.Println("❌ Permission NOT found in database")
		
		// Check similar permissions
		rows, err := conn.Query(ctx, `
			SELECT endpoint, method, description
			FROM permissions
			WHERE endpoint LIKE '%documents%'
			ORDER BY endpoint, method
		`)
		if err != nil {
			fmt.Printf("Failed to query: %v\n", err)
			return
		}
		defer rows.Close()
		
		fmt.Println("\nSimilar document-related permissions:")
		for rows.Next() {
			var endpoint, method, description string
			rows.Scan(&endpoint, &method, &description)
			fmt.Printf("  %-6s %-30s %s\n", method, endpoint, description)
		}
	}
}
