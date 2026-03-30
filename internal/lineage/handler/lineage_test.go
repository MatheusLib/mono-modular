package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"mono-modular/internal/lineage/repository"
)

type mockLineageService struct {
	recordID    uint64
	recordErr   error
	exportResult []repository.LineageEvent
	exportErr   error
}

func (m *mockLineageService) Record(_ context.Context, _ repository.LineageEvent) (uint64, error) {
	return m.recordID, m.recordErr
}

func (m *mockLineageService) ExportBySubject(_ context.Context, _ uint64) ([]repository.LineageEvent, error) {
	return m.exportResult, m.exportErr
}

func chiCtx(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// -- Record ----------------------------------------------------------------

func TestRecord_Success(t *testing.T) {
	svc := &mockLineageService{recordID: 7}
	h := LineageHandler{Service: svc}
	body := `{"subject_id":10,"operation":"COLLECT","source":"api","destination":"db","purpose":"analytics"}`
	req := httptest.NewRequest(http.MethodPost, "/lineage", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Record(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp map[string]uint64
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["id"] != 7 {
		t.Fatalf("expected id=7, got %d", resp["id"])
	}
}

func TestRecord_InvalidBody(t *testing.T) {
	svc := &mockLineageService{}
	h := LineageHandler{Service: svc}
	req := httptest.NewRequest(http.MethodPost, "/lineage", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.Record(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRecord_MissingFields(t *testing.T) {
	svc := &mockLineageService{}
	h := LineageHandler{Service: svc}
	body := `{"subject_id":10,"operation":"COLLECT"}`
	req := httptest.NewRequest(http.MethodPost, "/lineage", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Record(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestRecord_MissingSubjectID(t *testing.T) {
	svc := &mockLineageService{}
	h := LineageHandler{Service: svc}
	body := `{"operation":"COLLECT","source":"api","destination":"db","purpose":"analytics"}`
	req := httptest.NewRequest(http.MethodPost, "/lineage", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Record(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestRecord_ServiceError(t *testing.T) {
	svc := &mockLineageService{recordErr: errors.New("db fail")}
	h := LineageHandler{Service: svc}
	body := `{"subject_id":10,"operation":"COLLECT","source":"api","destination":"db","purpose":"analytics"}`
	req := httptest.NewRequest(http.MethodPost, "/lineage", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.Record(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// -- Export ----------------------------------------------------------------

func TestExport_Success(t *testing.T) {
	consentID := uint64(5)
	svc := &mockLineageService{exportResult: []repository.LineageEvent{
		{ID: 1, SubjectID: 10, Operation: "COLLECT", Source: "api", Destination: "db", Purpose: "analytics", ConsentID: &consentID, PayloadJSON: `{}`, CreatedAt: "2025-01-01 00:00:00"},
	}}
	h := LineageHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodGet, "/lineage/export/10", nil), "subject_id", "10")
	w := httptest.NewRecorder()
	h.Export(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []LineageEventResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].Operation != "COLLECT" {
		t.Fatalf("expected COLLECT, got %s", result[0].Operation)
	}
}

func TestExport_Empty(t *testing.T) {
	svc := &mockLineageService{exportResult: []repository.LineageEvent{}}
	h := LineageHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodGet, "/lineage/export/10", nil), "subject_id", "10")
	w := httptest.NewRecorder()
	h.Export(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestExport_InvalidSubjectID(t *testing.T) {
	svc := &mockLineageService{}
	h := LineageHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodGet, "/lineage/export/abc", nil), "subject_id", "abc")
	w := httptest.NewRecorder()
	h.Export(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestExport_ZeroSubjectID(t *testing.T) {
	svc := &mockLineageService{}
	h := LineageHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodGet, "/lineage/export/0", nil), "subject_id", "0")
	w := httptest.NewRecorder()
	h.Export(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestExport_ServiceError(t *testing.T) {
	svc := &mockLineageService{exportErr: errors.New("db down")}
	h := LineageHandler{Service: svc}
	req := chiCtx(httptest.NewRequest(http.MethodGet, "/lineage/export/10", nil), "subject_id", "10")
	w := httptest.NewRecorder()
	h.Export(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
