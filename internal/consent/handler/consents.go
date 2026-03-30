package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"mono-modular/internal/consent/repository"
	"mono-modular/internal/consent/service"
)

type Consent struct {
	ID        uint64  `json:"id"`
	UserID    uint64  `json:"user_id"`
	PolicyID  uint64  `json:"policy_id"`
	Purpose   string  `json:"purpose"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at,omitempty"`
	RevokedAt *string `json:"revoked_at,omitempty"`
}

type ConsentHandler struct {
	Service service.ConsentService
}

func NewConsentHandler(svc service.ConsentService) ConsentHandler {
	return ConsentHandler{Service: svc}
}

func (h ConsentHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 1 || parsed > 1000 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	consents, err := h.Service.ListConsents(ctx, limit)
	if err != nil {
		http.Error(w, "query error", http.StatusInternalServerError)
		return
	}

	resp := make([]Consent, 0, len(consents))
	for _, c := range consents {
		resp = append(resp, Consent{
			ID:       c.ID,
			UserID:   c.UserID,
			PolicyID: c.PolicyID,
			Purpose:  c.Purpose,
			Status:   c.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h ConsentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var body struct {
		UserID   uint64 `json:"user_id"`
		PolicyID uint64 `json:"policy_id"`
		Purpose  string `json:"purpose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.UserID == 0 || body.PolicyID == 0 || body.Purpose == "" {
		http.Error(w, "missing fields", http.StatusUnprocessableEntity)
		return
	}

	created, err := h.Service.CreateConsent(ctx, repository.Consent{
		UserID:   body.UserID,
		PolicyID: body.PolicyID,
		Purpose:  body.Purpose,
	})
	if err != nil {
		http.Error(w, "create error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(Consent{
		ID:       created.ID,
		UserID:   created.UserID,
		PolicyID: created.PolicyID,
		Purpose:  created.Purpose,
		Status:   created.Status,
	})
}

func (h ConsentHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	raw := chi.URLParam(r, "document_id")
	docID, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || docID == 0 {
		http.Error(w, "invalid document_id", http.StatusBadRequest)
		return
	}

	if err := h.Service.RevokeConsent(ctx, docID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "revoke error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
