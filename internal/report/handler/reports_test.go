package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mono-modular/internal/report/repository"
)

type mockReportService struct {
	listResult []repository.ConsentReport
	listErr    error
}

func (m *mockReportService) ListConsents(_ context.Context, _ *uint64, _ int) ([]repository.ConsentReport, error) {
	return m.listResult, m.listErr
}

func TestReportListConsents_Success(t *testing.T) {
	svc := &mockReportService{listResult: []repository.ConsentReport{
		{ID: 1, UserID: 10, PolicyID: 2, Purpose: "x", Status: "active"},
	}}
	h := ReportHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/reports/consents", nil)
	w := httptest.NewRecorder()
	h.ListConsents(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReportListConsents_Empty(t *testing.T) {
	svc := &mockReportService{listResult: []repository.ConsentReport{}}
	h := ReportHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/reports/consents", nil)
	w := httptest.NewRecorder()
	h.ListConsents(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReportListConsents_ServiceError(t *testing.T) {
	svc := &mockReportService{listErr: errors.New("db down")}
	h := ReportHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/reports/consents", nil)
	w := httptest.NewRecorder()
	h.ListConsents(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestReportListConsents_WithUserID(t *testing.T) {
	svc := &mockReportService{listResult: []repository.ConsentReport{}}
	h := ReportHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/reports/consents?user_id=5", nil)
	w := httptest.NewRecorder()
	h.ListConsents(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReportListConsents_InvalidUserID(t *testing.T) {
	svc := &mockReportService{}
	h := ReportHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/reports/consents?user_id=abc", nil)
	w := httptest.NewRecorder()
	h.ListConsents(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReportListConsents_InvalidLimit(t *testing.T) {
	svc := &mockReportService{}
	h := ReportHandler{Service: svc}
	req := httptest.NewRequest(http.MethodGet, "/reports/consents?limit=bad", nil)
	w := httptest.NewRecorder()
	h.ListConsents(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
