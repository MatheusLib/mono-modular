package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mono-modular/internal/policy/repository"
)

type mockPolicyService struct {
	listResult   []repository.Policy
	listErr      error
	createResult *repository.Policy
	createErr    error
}

func (m *mockPolicyService) ListPolicies(_ context.Context, _ int) ([]repository.Policy, error) {
	return m.listResult, m.listErr
}
func (m *mockPolicyService) CreatePolicy(_ context.Context, p repository.Policy) (*repository.Policy, error) {
	return m.createResult, m.createErr
}

func TestPolicyList_Success(t *testing.T) {
	svc := &mockPolicyService{listResult: []repository.Policy{
		{ID: 1, Version: "v1", ContentHash: "abc"},
	}}
	h := PolicyHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/policies", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []Policy
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestPolicyList_Empty(t *testing.T) {
	svc := &mockPolicyService{listResult: []repository.Policy{}}
	h := PolicyHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/policies", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPolicyList_ServiceError(t *testing.T) {
	svc := &mockPolicyService{listErr: errors.New("db down")}
	h := PolicyHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/policies", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestPolicyList_InvalidLimit(t *testing.T) {
	svc := &mockPolicyService{}
	h := PolicyHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/policies?limit=bad", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPolicyCreate_Success(t *testing.T) {
	created := &repository.Policy{ID: 3, Version: "v2", ContentHash: "hash"}
	svc := &mockPolicyService{createResult: created}
	h := PolicyHandler{Service: svc}
	body := `{"version":"v2","content_hash":"hash"}`
	req := httptest.NewRequest(http.MethodPost, "/policies", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestPolicyCreate_InvalidBody(t *testing.T) {
	svc := &mockPolicyService{}
	h := PolicyHandler{Service: svc}
	req := httptest.NewRequest(http.MethodPost, "/policies", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPolicyCreate_MissingFields(t *testing.T) {
	svc := &mockPolicyService{}
	h := PolicyHandler{Service: svc}
	body := `{"version":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/policies", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPolicyCreate_ServiceError(t *testing.T) {
	svc := &mockPolicyService{createErr: errors.New("fail")}
	h := PolicyHandler{Service: svc}
	body := `{"version":"v1","content_hash":"h"}`
	req := httptest.NewRequest(http.MethodPost, "/policies", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
