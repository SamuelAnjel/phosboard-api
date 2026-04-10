package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password"
	
	// Try to hash with default cost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}
	
	fmt.Printf("Generated hash: %s\n", hash)
	fmt.Printf("Hash length: %d\n", len(hash))
	
	// Try to compare with a known hash from DB
	dbHash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4KoEa3Ro9llC/.og/at2.uheWG/igi"
	err = bcrypt.CompareHashAndPassword([]byte(dbHash), []byte(password))
	if err != nil {
		fmt.Printf("Compare error: %v\n", err)
	} else {
		fmt.Println("✅ Hash matches password!")
	}
}
