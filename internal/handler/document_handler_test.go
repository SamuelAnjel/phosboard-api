package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrack_URLValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       TrackRequest
		wantStatus int
	}{
		{
			name:       "missing url",
			body:       TrackRequest{URL: ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid url format",
			body:       TrackRequest{URL: "not-a-url"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid url without scheme",
			body:       TrackRequest{URL: "example.com"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &DocumentHandler{}

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/documents/track", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Track(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestTrack_InvalidMethods(t *testing.T) {
	handler := &DocumentHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/documents/track", nil)
	rec := httptest.NewRecorder()

	handler.Track(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
