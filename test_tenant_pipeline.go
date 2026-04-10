package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	baseURL := "https://api-backend-wa2skw4ruq-ue.a.run.app"
	
	// 1. Login as tenant-admin
	fmt.Println("1. Logging in as tenant-admin...")
	loginData := map[string]string{
		"email":    "tenant.admin@example.com",
		"password": "password",
	}
	
	loginJSON, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	token, ok := result["token"].(string)
	if !ok {
		fmt.Printf("Raw response: %s\n", string(body))
		fmt.Println("Could not extract token from response")
		return
	}
	
	fmt.Printf("✅ Login successful. Token: %s...\n", token[:50])
	
	// 2. Test protected endpoint (should work with tenant context)
	fmt.Println("\n2. Testing protected endpoint with tenant context...")
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/documents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Protected request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		fmt.Println("✅ Documents endpoint accessible")
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}
	
	// 3. Test creating a document
	fmt.Println("\n3. Testing document creation...")
	docData := map[string]interface{}{
		"url":       "https://example.com/test-document-rbac",
		"source_id": "test-source-rbac",
		"priority":  1,
	}
	
	docJSON, _ := json.Marshal(docData)
	req, _ = http.NewRequest("POST", baseURL+"/api/v1/documents/track", bytes.NewBuffer(docJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Document creation failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Println("✅ Document creation successful!")
		fmt.Printf("Response: %s\n", string(body))
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}
	
	// 4. Test unauthorized endpoint (should fail)
	fmt.Println("\n4. Testing unauthorized endpoint (should fail)...")
	req, _ = http.NewRequest("POST", baseURL+"/api/v1/tenants", bytes.NewBuffer([]byte(`{"name":"Test Tenant"}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d (expected 403 Forbidden)\n", resp.StatusCode)
	if resp.StatusCode == 403 {
		fmt.Println("✅ RBAC correctly blocking unauthorized action")
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}
}
