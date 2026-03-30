package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRecord_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO data_lineage").
		WithArgs(uint64(10), "COLLECT", "api", "db", "analytics", nil, `{"key":"val"}`).
		WillReturnResult(sqlmock.NewResult(7, 1))

	repo := NewLineageRepository(db)
	id, err := repo.Record(context.Background(), LineageEvent{
		SubjectID:   10,
		Operation:   "COLLECT",
		Source:      "api",
		Destination: "db",
		Purpose:     "analytics",
		PayloadJSON: `{"key":"val"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 7 {
		t.Fatalf("expected ID=7, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestRecord_ExecError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("INSERT INTO data_lineage").WillReturnError(errors.New("insert fail"))
	repo := NewLineageRepository(db)
	_, err := repo.Record(context.Background(), LineageEvent{SubjectID: 1, Operation: "COLLECT", Source: "a", Destination: "b", Purpose: "c"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRecord_LastInsertIdError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("INSERT INTO data_lineage").
		WillReturnResult(sqlmock.NewErrorResult(errors.New("last id fail")))
	repo := NewLineageRepository(db)
	_, err := repo.Record(context.Background(), LineageEvent{SubjectID: 1, Operation: "COLLECT", Source: "a", Destination: "b", Purpose: "c"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListBySubject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	consentID := uint64(5)
	rows := sqlmock.NewRows([]string{"id", "subject_id", "operation", "source", "destination", "purpose", "consent_id", "payload_json", "created_at"}).
		AddRow(1, 10, "COLLECT", "api", "db", "analytics", consentID, `{}`, "2025-01-01 00:00:00").
		AddRow(2, 10, "SHARE", "db", "partner", "marketing", nil, `{}`, "2025-01-02 00:00:00")
	mock.ExpectQuery("SELECT id, subject_id").WillReturnRows(rows)

	repo := NewLineageRepository(db)
	events, err := repo.ListBySubject(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2, got %d", len(events))
	}
	if events[0].Operation != "COLLECT" {
		t.Fatalf("expected COLLECT, got %s", events[0].Operation)
	}
	if events[1].ConsentID != nil {
		t.Fatalf("expected nil consent_id for second event")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestListBySubject_Empty(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "subject_id", "operation", "source", "destination", "purpose", "consent_id", "payload_json", "created_at"})
	mock.ExpectQuery("SELECT id, subject_id").WillReturnRows(rows)

	repo := NewLineageRepository(db)
	events, err := repo.ListBySubject(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0, got %d", len(events))
	}
}

func TestListBySubject_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT id, subject_id").WillReturnError(errors.New("db down"))
	repo := NewLineageRepository(db)
	_, err := repo.ListBySubject(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListBySubject_RowsErr(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "subject_id", "operation", "source", "destination", "purpose", "consent_id", "payload_json", "created_at"}).
		AddRow(1, 10, "COLLECT", "api", "db", "analytics", nil, `{}`, "2025-01-01 00:00:00").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery("SELECT id, subject_id").WillReturnRows(rows)
	repo := NewLineageRepository(db)
	_, err := repo.ListBySubject(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
