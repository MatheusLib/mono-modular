package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"mono-modular/internal/lineage/repository"
	"mono-modular/internal/lineage/service"
)

type LineageEventResponse struct {
	ID          uint64  `json:"id"`
	SubjectID   uint64  `json:"subject_id"`
	Operation   string  `json:"operation"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Purpose     string  `json:"purpose"`
	ConsentID   *uint64 `json:"consent_id,omitempty"`
	PayloadJSON string  `json:"payload_json"`
	CreatedAt   string  `json:"created_at,omitempty"`
}

type LineageHandler struct {
	Service service.LineageService
}

func NewLineageHandler(svc service.LineageService) LineageHandler {
	return LineageHandler{Service: svc}
}

func (h LineageHandler) Record(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var body struct {
		SubjectID   uint64  `json:"subject_id"`
		Operation   string  `json:"operation"`
		Source      string  `json:"source"`
		Destination string  `json:"destination"`
		Purpose     string  `json:"purpose"`
		ConsentID   *uint64 `json:"consent_id,omitempty"`
		PayloadJSON string  `json:"payload_json,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.SubjectID == 0 || body.Operation == "" || body.Source == "" || body.Destination == "" || body.Purpose == "" {
		http.Error(w, "missing fields", http.StatusUnprocessableEntity)
		return
	}

	id, err := h.Service.Record(ctx, repository.LineageEvent{
		SubjectID:   body.SubjectID,
		Operation:   body.Operation,
		Source:      body.Source,
		Destination: body.Destination,
		Purpose:     body.Purpose,
		ConsentID:   body.ConsentID,
		PayloadJSON: body.PayloadJSON,
	})
	if err != nil {
		http.Error(w, "record error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]uint64{"id": id})
}

func (h LineageHandler) Export(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	raw := chi.URLParam(r, "subject_id")
	subjectID, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || subjectID == 0 {
		http.Error(w, "invalid subject_id", http.StatusBadRequest)
		return
	}

	events, err := h.Service.ExportBySubject(ctx, subjectID)
	if err != nil {
		http.Error(w, "export error", http.StatusInternalServerError)
		return
	}

	resp := make([]LineageEventResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, LineageEventResponse{
			ID:          e.ID,
			SubjectID:   e.SubjectID,
			Operation:   e.Operation,
			Source:      e.Source,
			Destination: e.Destination,
			Purpose:     e.Purpose,
			ConsentID:   e.ConsentID,
			PayloadJSON: e.PayloadJSON,
			CreatedAt:   e.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
