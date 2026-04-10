package main

import (
	"context"
	"fmt"
	"os"
	"time"
	
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
	
	// Add missing permission
	_, err = conn.Exec(ctx, `
		INSERT INTO permissions (id, endpoint, method, description, created_at) VALUES
			('11111111-1111-1111-1111-111111111108', '/api/v1/documents/track', 'POST', 'Track new documents', $1)
		ON CONFLICT (id) DO UPDATE SET
			endpoint = EXCLUDED.endpoint,
			method = EXCLUDED.method,
			description = EXCLUDED.description
	`, time.Now())
	
	if err != nil {
		fmt.Printf("Failed to add permission: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Added permission: POST /api/v1/documents/track")
	
	// Assign to tenant-admin role
	_, err = conn.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
			('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111108', $1)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, time.Now())
	
	if err != nil {
		fmt.Printf("Failed to assign permission to role: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Assigned permission to tenant-admin role")
	
	// Verify
	var count int
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = '22222222-2222-2222-2222-222222222222'
		AND p.endpoint = '/api/v1/documents/track'
		AND p.method = 'POST'
	`).Scan(&count)
	
	if err != nil {
		fmt.Printf("Failed to verify: %v\n", err)
	} else if count > 0 {
		fmt.Println("✅ Permission verified in database")
	} else {
		fmt.Println("❌ Permission not found after insertion")
	}
}
