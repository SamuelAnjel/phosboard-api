package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"phosboard/backend/internal/models"
)

func TestInsertGlobalDocument_Error_InvalidPool(t *testing.T) {
	poolConfig, _ := pgxpool.ParseConfig("postgres://invalid")
	poolConfig.ConnConfig.Host = "invalid-host"
	pool, _ := pgxpool.NewWithConfig(context.Background(), poolConfig)
	repo := &PostgresDocumentRepository{pool: pool}

	doc := models.GlobalDocument{
		SourceID:    "source-1",
		Title:       "Test",
		URL:         "https://example.com",
		ContentText: "content",
	}

	_, err := repo.InsertGlobalDocument(context.Background(), doc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLinkDocumentToTenant_Error_InvalidPool(t *testing.T) {
	poolConfig, _ := pgxpool.ParseConfig("postgres://invalid")
	poolConfig.ConnConfig.Host = "invalid-host"
	pool, _ := pgxpool.NewWithConfig(context.Background(), poolConfig)
	repo := &PostgresDocumentRepository{pool: pool}

	err := repo.LinkDocumentToTenant(context.Background(), "tenant-1", "doc-1", []string{"keyword"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetLatestByTenant_Error_InvalidPool(t *testing.T) {
	poolConfig, _ := pgxpool.ParseConfig("postgres://invalid")
	poolConfig.ConnConfig.Host = "invalid-host"
	pool, _ := pgxpool.NewWithConfig(context.Background(), poolConfig)
	repo := &PostgresDocumentRepository{pool: pool}

	_, err := repo.GetLatestByTenant(context.Background(), "tenant-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDocumentRepository_Interface(t *testing.T) {
	var _ DocumentRepository = (*PostgresDocumentRepository)(nil)
}
