package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	client *storage.Client
	bucket string
}

func NewGCSStorage(ctx context.Context, bucket string) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GCS client: %w", err)
	}

	return &GCSStorage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *GCSStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	obj := s.client.Bucket(s.bucket).Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("get object from GCS: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}

	return data, nil
}

func (s *GCSStorage) PutObject(ctx context.Context, key string, data []byte) error {
	obj := s.client.Bucket(s.bucket).Object(key)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/json"

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("put object to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close GCS writer: %w", err)
	}

	return nil
}

func (s *GCSStorage) Bucket() string {
	return s.bucket
}
