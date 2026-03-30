package service

import (
	"context"

	"mono-modular/internal/consent/repository"
)

type ConsentService interface {
	ListConsents(ctx context.Context, limit int) ([]repository.Consent, error)
	CreateConsent(ctx context.Context, c repository.Consent) (*repository.Consent, error)
	RevokeConsent(ctx context.Context, documentID uint64) error
}

type consentService struct {
	repo repository.ConsentRepository
}

func NewConsentService(repo repository.ConsentRepository) ConsentService {
	return &consentService{repo: repo}
}

func (s *consentService) ListConsents(ctx context.Context, limit int) ([]repository.Consent, error) {
	return s.repo.List(ctx, limit)
}

func (s *consentService) CreateConsent(ctx context.Context, c repository.Consent) (*repository.Consent, error) {
	return s.repo.Create(ctx, c)
}

func (s *consentService) RevokeConsent(ctx context.Context, documentID uint64) error {
	return s.repo.Revoke(ctx, documentID)
}
