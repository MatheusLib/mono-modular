package service

import (
	"context"
	"errors"
	"testing"

	"mono-modular/internal/policy/repository"
)

type mockPolicyRepo struct {
	listResult   []repository.Policy
	listErr      error
	createResult *repository.Policy
	createErr    error
}

func (m *mockPolicyRepo) List(_ context.Context, _ int) ([]repository.Policy, error) {
	return m.listResult, m.listErr
}
func (m *mockPolicyRepo) Create(_ context.Context, p repository.Policy) (*repository.Policy, error) {
	return m.createResult, m.createErr
}

func TestListPolicies_Success(t *testing.T) {
	repo := &mockPolicyRepo{listResult: []repository.Policy{{ID: 1, Version: "v1"}}}
	svc := NewPolicyService(repo)
	result, err := svc.ListPolicies(context.Background(), 10)
	if err != nil || len(result) != 1 {
		t.Fatalf("unexpected: err=%v len=%d", err, len(result))
	}
}

func TestListPolicies_Error(t *testing.T) {
	repo := &mockPolicyRepo{listErr: errors.New("db error")}
	svc := NewPolicyService(repo)
	_, err := svc.ListPolicies(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreatePolicy_Success(t *testing.T) {
	expected := &repository.Policy{ID: 3, Version: "v2", ContentHash: "abc"}
	repo := &mockPolicyRepo{createResult: expected}
	svc := NewPolicyService(repo)
	p, err := svc.CreatePolicy(context.Background(), repository.Policy{Version: "v2", ContentHash: "abc"})
	if err != nil || p.ID != 3 {
		t.Fatalf("unexpected: err=%v id=%d", err, p.ID)
	}
}

func TestCreatePolicy_Error(t *testing.T) {
	repo := &mockPolicyRepo{createErr: errors.New("fail")}
	svc := NewPolicyService(repo)
	_, err := svc.CreatePolicy(context.Background(), repository.Policy{})
	if err == nil {
		t.Fatal("expected error")
	}
}
