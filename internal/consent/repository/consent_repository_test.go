package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestList_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "policy_id", "purpose", "status"}).
		AddRow(1, 10, 2, "marketing", "active").
		AddRow(2, 11, 3, "analytics", "revoked")
	mock.ExpectQuery("SELECT id, user_id").WillReturnRows(rows)

	repo := NewConsentRepository(db)
	consents, err := repo.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(consents) != 2 {
		t.Fatalf("expected 2, got %d", len(consents))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestList_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT id, user_id").WillReturnError(errors.New("db down"))
	repo := NewConsentRepository(db)
	_, err := repo.List(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestList_RowsErr(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "user_id", "policy_id", "purpose", "status"}).
		AddRow(1, 10, 2, "marketing", "active").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery("SELECT id, user_id").WillReturnRows(rows)
	repo := NewConsentRepository(db)
	_, err := repo.List(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO consents").
		WithArgs(uint64(10), uint64(2), "marketing").
		WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectExec("INSERT INTO audit_events").
		WithArgs(int64(5), `{"consent_id":5}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewConsentRepository(db)
	c, err := repo.Create(context.Background(), Consent{UserID: 10, PolicyID: 2, Purpose: "marketing"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.ID != 5 {
		t.Fatalf("expected ID=5, got %d", c.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestCreate_BeginError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin().WillReturnError(errors.New("begin fail"))
	repo := NewConsentRepository(db)
	_, err := repo.Create(context.Background(), Consent{UserID: 1, PolicyID: 1, Purpose: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_InsertError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO consents").WillReturnError(errors.New("insert fail"))
	mock.ExpectRollback()
	repo := NewConsentRepository(db)
	_, err := repo.Create(context.Background(), Consent{UserID: 1, PolicyID: 1, Purpose: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_AuditError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO consents").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO audit_events").WillReturnError(errors.New("audit fail"))
	mock.ExpectRollback()
	repo := NewConsentRepository(db)
	_, err := repo.Create(context.Background(), Consent{UserID: 1, PolicyID: 1, Purpose: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRevoke_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE consents").
		WithArgs(uint64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO audit_events").
		WithArgs(uint64(3), `{"consent_id":3}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewConsentRepository(db)
	if err := repo.Revoke(context.Background(), 3); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestRevoke_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE consents").
		WithArgs(uint64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	repo := NewConsentRepository(db)
	err := repo.Revoke(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestRevoke_UpdateError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE consents").WillReturnError(errors.New("update fail"))
	mock.ExpectRollback()
	repo := NewConsentRepository(db)
	if err := repo.Revoke(context.Background(), 1); err == nil {
		t.Fatal("expected error")
	}
}
