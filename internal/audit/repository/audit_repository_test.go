package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAuditList_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "event_type", "entity_type", "entity_id", "payload_json"}).
		AddRow(1, "ConsentCreated", "consent", 1, `{"consent_id":1}`).
		AddRow(2, "ConsentRevoked", "consent", 1, `{"consent_id":1}`)
	mock.ExpectQuery("SELECT id, event_type").WillReturnRows(rows)
	repo := NewAuditRepository(db)
	events, err := repo.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2, got %d", len(events))
	}
}

func TestAuditList_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT id, event_type").WillReturnError(errors.New("db down"))
	repo := NewAuditRepository(db)
	_, err := repo.List(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuditList_RowsErr(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "event_type", "entity_type", "entity_id", "payload_json"}).
		AddRow(1, "ConsentCreated", "consent", 1, `{}`).
		RowError(0, errors.New("row error"))
	mock.ExpectQuery("SELECT id, event_type").WillReturnRows(rows)
	repo := NewAuditRepository(db)
	_, err := repo.List(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
