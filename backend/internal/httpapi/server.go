package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"hublead-attio-integration/backend/internal/attio"
	"hublead-attio-integration/backend/internal/domain"
	"hublead-attio-integration/backend/internal/service"
)

type App interface {
	Lookup(ctx context.Context, linkedinURL string, rawDomain string) (domain.CompanySummary, error)
	Sync(ctx context.Context, input domain.ExtractedCompany) (domain.CompanySummary, error)
}

type HealthStore interface {
	Ping(ctx context.Context) error
}

type Server struct {
	app    App
	health HealthStore
	logger *slog.Logger
}

func NewServer(app App, health HealthStore, allowedOrigins []string, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{app: app, health: health, logger: logger}

	r := chi.NewRouter()
	r.Use(requestLogger(logger))
	r.Use(cors(allowedOrigins))
	r.Get("/healthz", s.healthz)
	r.Get("/api/companies/lookup", s.lookupCompany)
	r.Post("/api/companies/sync", s.syncCompany)
	return r
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if err := s.health.Ping(r.Context()); err != nil {
		s.writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}

func (s *Server) lookupCompany(w http.ResponseWriter, r *http.Request) {
	company, err := s.app.Lookup(r.Context(), r.URL.Query().Get("linkedin_url"), r.URL.Query().Get("domain"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"status": "not_found"})
			return
		}
		s.writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "found", "company": company})
}

func (s *Server) syncCompany(w http.ResponseWriter, r *http.Request) {
	var input domain.ExtractedCompany
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	company, err := s.app.Sync(r.Context(), input)
	if err != nil {
		s.writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "synced", "company": company})
}

func (s *Server) writeMappedError(w http.ResponseWriter, err error) {
	var validation service.ValidationError
	if errors.As(err, &validation) {
		s.writeError(w, http.StatusUnprocessableEntity, validation.Message)
		return
	}

	if attio.IsAuthError(err) {
		s.writeError(w, http.StatusUnauthorized, "Attio credentials were rejected")
		return
	}

	if attio.IsAPIError(err) {
		s.writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	s.logger.Error("unexpected request error", "error", err)
	s.writeError(w, http.StatusInternalServerError, "unexpected backend error")
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func cors(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if isAllowedOrigin(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Hublead-Client")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed map[string]struct{}) bool {
	if origin == "" {
		return false
	}
	if _, ok := allowed[origin]; ok {
		return true
	}
	return strings.HasPrefix(origin, "chrome-extension://")
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"origin", r.Header.Get("Origin"),
				"hublead_client", r.Header.Get("X-Hublead-Client"),
				"user_agent", r.Header.Get("User-Agent"),
			)
			next.ServeHTTP(w, r)
		})
	}
}
