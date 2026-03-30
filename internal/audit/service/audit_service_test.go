package service

import (
	"context"
	"errors"
	"testing"

	"mono-modular/internal/audit/repository"
)

type mockAuditRepo struct {
	listResult []repository.AuditEvent
	listErr    error
}

func (m *mockAuditRepo) List(_ context.Context, _ int) ([]repository.AuditEvent, error) {
	return m.listResult, m.listErr
}

func TestListEvents_Success(t *testing.T) {
	repo := &mockAuditRepo{listResult: []repository.AuditEvent{{ID: 1}}}
	svc := NewAuditService(repo)
	result, err := svc.ListEvents(context.Background(), 10)
	if err != nil || len(result) != 1 {
		t.Fatalf("unexpected: err=%v len=%d", err, len(result))
	}
}

func TestListEvents_Error(t *testing.T) {
	repo := &mockAuditRepo{listErr: errors.New("db error")}
	svc := NewAuditService(repo)
	_, err := svc.ListEvents(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
