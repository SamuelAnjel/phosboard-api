package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	ctx := context.Background()

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}

	client, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	buckets := []string{
		"raw-html",
		"social-payloads",
		"parquet-files",
	}

	for _, bucketName := range buckets {
		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			log.Printf("Error checking bucket %s: %v", bucketName, err)
			continue
		}
		if !exists {
			err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("Failed to create bucket %s: %v", bucketName, err)
			} else {
				log.Printf("Created bucket: %s", bucketName)
			}
		} else {
			log.Printf("Bucket already exists: %s", bucketName)
		}
	}

	fmt.Println("Done!")
}
