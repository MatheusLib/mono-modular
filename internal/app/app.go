package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mono-modular/internal/consent/handler"
	"mono-modular/internal/consent/repository"
	"mono-modular/internal/consent/service"
	"mono-modular/internal/middleware"
	audithandler "mono-modular/internal/audit/handler"
	auditrepo "mono-modular/internal/audit/repository"
	auditservice "mono-modular/internal/audit/service"
	policyhandler "mono-modular/internal/policy/handler"
	policyrepo "mono-modular/internal/policy/repository"
	policyservice "mono-modular/internal/policy/service"
	reporthandler "mono-modular/internal/report/handler"
	reportrepo "mono-modular/internal/report/repository"
	reportservice "mono-modular/internal/report/service"
	lineagehandler "mono-modular/internal/lineage/handler"
	lineagerepo "mono-modular/internal/lineage/repository"
	lineageservice "mono-modular/internal/lineage/service"
)

func NewRouter(db *sql.DB) http.Handler {
	repo := repository.NewConsentRepository(db)
	svc := service.NewConsentService(repo)
	h := handler.NewConsentHandler(svc)

	policyRepo := policyrepo.NewPolicyRepository(db)
	policySvc := policyservice.NewPolicyService(policyRepo)
	policyHandler := policyhandler.NewPolicyHandler(policySvc)

	auditRepo := auditrepo.NewAuditRepository(db)
	auditSvc := auditservice.NewAuditService(auditRepo)
	auditHandler := audithandler.NewAuditHandler(auditSvc)

	reportRepo := reportrepo.NewReportRepository(db)
	reportSvc := reportservice.NewReportService(reportRepo)
	reportHandler := reporthandler.NewReportHandler(reportSvc)

	lineageRepo := lineagerepo.NewLineageRepository(db)
	lineageSvc := lineageservice.NewLineageService(lineageRepo)
	lineageH := lineagehandler.NewLineageHandler(lineageSvc)

	r := chi.NewRouter()

	logger := middleware.NewLogger()
	r.Use(middleware.RequestID)
	r.Use(middleware.Tracing)
	r.Use(middleware.Logging(logger))

	r.Get("/health", Health)
	r.Get("/consents", h.List)
	r.Post("/consents", h.Create)
	r.Patch("/consents/{document_id}/revoke", h.Revoke)
	r.Get("/policies", policyHandler.List)
	r.Post("/policies", policyHandler.Create)
	r.Get("/audit/events", auditHandler.List)
	r.Get("/reports/consents", reportHandler.ListConsents)
	r.Post("/lineage", lineageH.Record)
	r.Get("/lineage/export/{subject_id}", lineageH.Export)

	return r
}
