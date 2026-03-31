package models

import "time"

type TenantConcept struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ConceptTerm string    `json:"concept_term"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateConceptRequest struct {
	ConceptTerm string `json:"concept_term"`
}
