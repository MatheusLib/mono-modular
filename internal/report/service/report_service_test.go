package service

import (
	"context"
	"errors"
	"testing"

	"mono-modular/internal/report/repository"
)

type mockReportRepo struct {
	listResult []repository.ConsentReport
	listErr    error
}

func (m *mockReportRepo) ListConsents(_ context.Context, _ *uint64, _ int) ([]repository.ConsentReport, error) {
	return m.listResult, m.listErr
}

func TestReportListConsents_Success(t *testing.T) {
	repo := &mockReportRepo{listResult: []repository.ConsentReport{{ID: 1}}}
	svc := NewReportService(repo)
	result, err := svc.ListConsents(context.Background(), nil, 10)
	if err != nil || len(result) != 1 {
		t.Fatalf("unexpected: err=%v len=%d", err, len(result))
	}
}

func TestReportListConsents_Error(t *testing.T) {
	repo := &mockReportRepo{listErr: errors.New("db error")}
	svc := NewReportService(repo)
	_, err := svc.ListConsents(context.Background(), nil, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
