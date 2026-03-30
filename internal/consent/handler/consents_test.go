package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"mono-modular/internal/consent/repository"
)

// mockConsentService implements service.ConsentService
type mockConsentService struct {
	listResult   []repository.Consent
	listErr      error
	createResult *repository.Consent
	createErr    error
	revokeErr    error
}

func (m *mockConsentService) ListConsents(_ context.Context, _ int) ([]repository.Consent, error) {
	return m.listResult, m.listErr
}
func (m *mockConsentService) CreateConsent(_ context.Context, _ repository.Consent) (*repository.Consent, error) {
	return m.createResult, m.createErr
}
func (m *mockConsentService) RevokeConsent(_ context.Context, _ uint64) error {
	return m.revokeErr
}

func chiCtx(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestConsentList_Success(t *testing.T) {
	svc := &mockConsentService{listResult: []repository.Consent{
		{ID: 1, UserID: 10, PolicyID: 2, Purpose: "marketing", Status: "active"},
	}}
	h := ConsentHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/consents", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []Consent
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestConsentList_Empty(t *testing.T) {
	svc := &mockConsentService{listResult: []repository.Consent{}}
	h := ConsentHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/consents", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestConsentList_ServiceError(t *testing.T) {
	svc := &mockConsentService{listErr: errors.New("db down")}
	h := ConsentHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/consents", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestConsentList_InvalidLimit(t *testing.T) {
	svc := &mockConsentService{}
	h := ConsentHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/consents?limit=abc", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestConsentCreate_Success(t *testing.T) {
	created := &repository.Consent{ID: 5, UserID: 1, PolicyID: 2, Purpose: "marketing", Status: "active"}
	svc := &mockConsentService{createResult: created}
	h := ConsentHandler{Service: svc}
	body := `{"user_id":1,"policy_id":2,"purpose":"marketing"}`
	req := httptest.NewRequest(http.MethodPost, "/consents", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestConsentCreate_InvalidBody(t *testing.T) {
	svc := &mockConsentService{}
	h := ConsentHandler{Service: svc}
	req := httptest.NewRequest(http.MethodPost, "/consents", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestConsentCreate_MissingFields(t *testing.T) {
	svc := &mockConsentService{}
	h := ConsentHandler{Service: svc}
	body := `{"user_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/consents", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestConsentCreate_ServiceError(t *testing.T) {
	svc := &mockConsentService{createErr: errors.New("fail")}
	h := ConsentHandler{Service: svc}
	body := `{"user_id":1,"policy_id":2,"purpose":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/consents", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// ── Revoke ────────────────────────────────────────────────────────────────────

func TestConsentRevoke_Success(t *testing.T) {
	svc := &mockConsentService{}
	h := ConsentHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodPatch, "/consents/3/revoke", nil), "document_id", "3")
	w := httptest.NewRecorder()
	h.Revoke(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestConsentRevoke_InvalidDocumentID(t *testing.T) {
	svc := &mockConsentService{}
	h := ConsentHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodPatch, "/consents/abc/revoke", nil), "document_id", "abc")
	w := httptest.NewRecorder()
	h.Revoke(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestConsentRevoke_NotFound(t *testing.T) {
	svc := &mockConsentService{revokeErr: sql.ErrNoRows}
	h := ConsentHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodPatch, "/consents/99/revoke", nil), "document_id", "99")
	w := httptest.NewRecorder()
	h.Revoke(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestConsentRevoke_ServiceError(t *testing.T) {
	svc := &mockConsentService{revokeErr: errors.New("tx error")}
	h := ConsentHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodPatch, "/consents/1/revoke", nil), "document_id", "1")
	w := httptest.NewRecorder()
	h.Revoke(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
