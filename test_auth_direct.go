package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	
	"phosboard/backend/internal/repository"
)

func main() {
	dbURL := "postgresql://postgres.ohrmoiplfblbzstpgpxn:&2s9d-3cXALSPtd@aws-1-us-east-1.pooler.supabase.com:5432/postgres"
	
	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed to create pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	
	userRepo := repository.NewUserRepository(pool)
	
	fmt.Println("Testing ValidateCredentials directly...")
	
	user, err := userRepo.ValidateCredentials(ctx, "admin@phosboard.cl", "password")
	if err != nil {
		fmt.Printf("❌ ValidateCredentials failed: %v\n", err)
		
		// Test GetUserByEmail directly
		user2, err2 := userRepo.GetUserByEmail(ctx, "admin@phosboard.cl")
		if err2 != nil {
			fmt.Printf("❌ GetUserByEmail also failed: %v\n", err2)
		} else {
			fmt.Printf("✅ GetUserByEmail succeeded: %s\n", user2.Email)
			fmt.Printf("  Password hash: %s...\n", user2.PasswordHash[:30])
			
			// Test bcrypt directly
			err3 := bcrypt.CompareHashAndPassword([]byte(user2.PasswordHash), []byte("password"))
			if err3 != nil {
				fmt.Printf("❌ bcrypt.CompareHashAndPassword failed: %v\n", err3)
			} else {
				fmt.Println("✅ bcrypt.CompareHashAndPassword succeeded!")
			}
		}
	} else {
		fmt.Printf("✅ ValidateCredentials succeeded: %s\n", user.Email)
	}
}
