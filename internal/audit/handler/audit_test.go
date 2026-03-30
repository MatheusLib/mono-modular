package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mono-modular/internal/audit/repository"
)

type mockAuditService struct {
	listResult []repository.AuditEvent
	listErr    error
}

func (m *mockAuditService) ListEvents(_ context.Context, _ int) ([]repository.AuditEvent, error) {
	return m.listResult, m.listErr
}

func TestAuditList_Success(t *testing.T) {
	svc := &mockAuditService{listResult: []repository.AuditEvent{
		{ID: 1, EventType: "ConsentCreated", EntityType: "consent", EntityID: 1, Payload: `{}`},
	}}
	h := AuditHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/audit/events", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []AuditEvent
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestAuditList_Empty(t *testing.T) {
	svc := &mockAuditService{listResult: []repository.AuditEvent{}}
	h := AuditHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/audit/events", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuditList_ServiceError(t *testing.T) {
	svc := &mockAuditService{listErr: errors.New("db down")}
	h := AuditHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/audit/events", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestAuditList_InvalidLimit(t *testing.T) {
	svc := &mockAuditService{}
	h := AuditHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/audit/events?limit=xyz", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
