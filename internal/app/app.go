package app

import (
	"database/sql"
	"net/http"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", Health)
	mux.HandleFunc("/consents", h.List)
	mux.HandleFunc("/policies", policyHandler.List)
	mux.HandleFunc("/audit/events", auditHandler.List)
	mux.HandleFunc("/reports/consents", reportHandler.ListConsents)

	logger := middleware.NewLogger()
	return middleware.RequestID(middleware.Logging(logger)(middleware.Tracing(mux)))
}
