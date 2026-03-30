package service

import (
	"context"
	"errors"
	"testing"

	"mono-modular/internal/consent/repository"
)

type mockConsentRepo struct {
	listResult   []repository.Consent
	listErr      error
	createResult *repository.Consent
	createErr    error
	revokeErr    error
}

func (m *mockConsentRepo) List(_ context.Context, _ int) ([]repository.Consent, error) {
	return m.listResult, m.listErr
}
func (m *mockConsentRepo) Create(_ context.Context, c repository.Consent) (*repository.Consent, error) {
	return m.createResult, m.createErr
}
func (m *mockConsentRepo) Revoke(_ context.Context, _ uint64) error {
	return m.revokeErr
}

func TestListConsents_Success(t *testing.T) {
	repo := &mockConsentRepo{listResult: []repository.Consent{{ID: 1}}}
	svc := NewConsentService(repo)
	result, err := svc.ListConsents(context.Background(), 10)
	if err != nil || len(result) != 1 {
		t.Fatalf("unexpected: err=%v len=%d", err, len(result))
	}
}

func TestListConsents_Error(t *testing.T) {
	repo := &mockConsentRepo{listErr: errors.New("db error")}
	svc := NewConsentService(repo)
	_, err := svc.ListConsents(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateConsent_Success(t *testing.T) {
	expected := &repository.Consent{ID: 5, UserID: 1, PolicyID: 2, Purpose: "x", Status: "active"}
	repo := &mockConsentRepo{createResult: expected}
	svc := NewConsentService(repo)
	c, err := svc.CreateConsent(context.Background(), repository.Consent{UserID: 1, PolicyID: 2, Purpose: "x"})
	if err != nil || c.ID != 5 {
		t.Fatalf("unexpected: err=%v id=%d", err, c.ID)
	}
}

func TestCreateConsent_Error(t *testing.T) {
	repo := &mockConsentRepo{createErr: errors.New("fail")}
	svc := NewConsentService(repo)
	_, err := svc.CreateConsent(context.Background(), repository.Consent{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRevokeConsent_Success(t *testing.T) {
	repo := &mockConsentRepo{}
	svc := NewConsentService(repo)
	if err := svc.RevokeConsent(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRevokeConsent_Error(t *testing.T) {
	repo := &mockConsentRepo{revokeErr: errors.New("not found")}
	svc := NewConsentService(repo)
	if err := svc.RevokeConsent(context.Background(), 99); err == nil {
		t.Fatal("expected error")
	}
}
