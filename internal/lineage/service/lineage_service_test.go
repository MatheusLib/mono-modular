package service

import (
	"context"
	"errors"
	"testing"

	"mono-modular/internal/lineage/repository"
)

type mockLineageRepo struct {
	recordID       uint64
	recordErr      error
	listResult     []repository.LineageEvent
	listErr        error
}

func (m *mockLineageRepo) Record(_ context.Context, _ repository.LineageEvent) (uint64, error) {
	return m.recordID, m.recordErr
}

func (m *mockLineageRepo) ListBySubject(_ context.Context, _ uint64) ([]repository.LineageEvent, error) {
	return m.listResult, m.listErr
}

func TestRecord_Success(t *testing.T) {
	repo := &mockLineageRepo{recordID: 42}
	svc := NewLineageService(repo)
	id, err := svc.Record(context.Background(), repository.LineageEvent{SubjectID: 1, Operation: "COLLECT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestRecord_Error(t *testing.T) {
	repo := &mockLineageRepo{recordErr: errors.New("fail")}
	svc := NewLineageService(repo)
	_, err := svc.Record(context.Background(), repository.LineageEvent{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExportBySubject_Success(t *testing.T) {
	repo := &mockLineageRepo{listResult: []repository.LineageEvent{
		{ID: 1, SubjectID: 10, Operation: "COLLECT"},
		{ID: 2, SubjectID: 10, Operation: "SHARE"},
	}}
	svc := NewLineageService(repo)
	events, err := svc.ExportBySubject(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2, got %d", len(events))
	}
}

func TestExportBySubject_Empty(t *testing.T) {
	repo := &mockLineageRepo{listResult: []repository.LineageEvent{}}
	svc := NewLineageService(repo)
	events, err := svc.ExportBySubject(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0, got %d", len(events))
	}
}

func TestExportBySubject_Error(t *testing.T) {
	repo := &mockLineageRepo{listErr: errors.New("db error")}
	svc := NewLineageService(repo)
	_, err := svc.ExportBySubject(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
