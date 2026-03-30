package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPolicyList_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "version", "content_hash"}).
		AddRow(1, "v1", "abc123").
		AddRow(2, "v2", "def456")
	mock.ExpectQuery("SELECT id, version").WillReturnRows(rows)
	repo := NewPolicyRepository(db)
	policies, err := repo.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(policies) != 2 {
		t.Fatalf("expected 2, got %d", len(policies))
	}
}

func TestPolicyList_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT id, version").WillReturnError(errors.New("db down"))
	repo := NewPolicyRepository(db)
	_, err := repo.List(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPolicyList_RowsErr(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "version", "content_hash"}).
		AddRow(1, "v1", "abc").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery("SELECT id, version").WillReturnRows(rows)
	repo := NewPolicyRepository(db)
	_, err := repo.List(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPolicyCreate_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("INSERT INTO policies").
		WithArgs("v3", "hash999").
		WillReturnResult(sqlmock.NewResult(7, 1))
	repo := NewPolicyRepository(db)
	p, err := repo.Create(context.Background(), Policy{Version: "v3", ContentHash: "hash999"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID != 7 {
		t.Fatalf("expected ID=7, got %d", p.ID)
	}
}

func TestPolicyCreate_InsertError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("INSERT INTO policies").WillReturnError(errors.New("fail"))
	repo := NewPolicyRepository(db)
	_, err := repo.Create(context.Background(), Policy{Version: "v1", ContentHash: "h"})
	if err == nil {
		t.Fatal("expected error")
	}
}
