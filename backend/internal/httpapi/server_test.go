package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hublead-attio-integration/backend/internal/attio"
	"hublead-attio-integration/backend/internal/domain"
	"hublead-attio-integration/backend/internal/service"
)

type fakeApp struct {
	lookupFn func(context.Context, string, string) (domain.CompanySummary, error)
	syncFn   func(context.Context, domain.ExtractedCompany) (domain.CompanySummary, error)
}

func (f fakeApp) Lookup(ctx context.Context, linkedinURL string, rawDomain string) (domain.CompanySummary, error) {
	return f.lookupFn(ctx, linkedinURL, rawDomain)
}

func (f fakeApp) Sync(ctx context.Context, input domain.ExtractedCompany) (domain.CompanySummary, error) {
	return f.syncFn(ctx, input)
}

type fakeHealth struct {
	err error
}

func (f fakeHealth) Ping(context.Context) error {
	return f.err
}

func TestLookupFound(t *testing.T) {
	summary := testCompany()
	handler := NewServer(fakeApp{
		lookupFn: func(context.Context, string, string) (domain.CompanySummary, error) {
			return summary, nil
		},
	}, fakeHealth{}, nil, nil)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/companies/lookup?linkedin_url=https://www.linkedin.com/company/attio", nil)
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "found" {
		t.Fatalf("status body = %#v", body)
	}
}

func TestLookupNotFound(t *testing.T) {
	handler := NewServer(fakeApp{
		lookupFn: func(context.Context, string, string) (domain.CompanySummary, error) {
			return domain.CompanySummary{}, service.ErrNotFound
		},
	}, fakeHealth{}, nil, nil)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/companies/lookup?linkedin_url=https://www.linkedin.com/company/attio", nil)
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d", resp.Code)
	}
}

func TestSyncValidation(t *testing.T) {
	handler := NewServer(fakeApp{
		syncFn: func(context.Context, domain.ExtractedCompany) (domain.CompanySummary, error) {
			return domain.CompanySummary{}, service.ValidationError{Message: "domain is required"}
		},
	}, fakeHealth{}, nil, nil)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/companies/sync", bytes.NewBufferString(`{"name":"Attio"}`))
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", resp.Code)
	}
}

func TestSyncSuccess(t *testing.T) {
	summary := testCompany()
	handler := NewServer(fakeApp{
		syncFn: func(context.Context, domain.ExtractedCompany) (domain.CompanySummary, error) {
			return summary, nil
		},
	}, fakeHealth{}, nil, nil)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/companies/sync", bytes.NewBufferString(`{"name":"Attio","linkedin_url":"https://www.linkedin.com/company/attio","domain":"attio.com"}`))
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
}

func TestAttioFailureMapsToBackendFailure(t *testing.T) {
	handler := NewServer(fakeApp{
		syncFn: func(context.Context, domain.ExtractedCompany) (domain.CompanySummary, error) {
			return domain.CompanySummary{}, attio.APIError{StatusCode: http.StatusBadGateway, Message: "Attio request failed"}
		},
	}, fakeHealth{}, nil, nil)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/companies/sync", bytes.NewBufferString(`{"name":"Attio","linkedin_url":"https://www.linkedin.com/company/attio","domain":"attio.com"}`))
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", resp.Code)
	}
}

func TestUnexpectedFailureMapsToInternalServerError(t *testing.T) {
	handler := NewServer(fakeApp{
		syncFn: func(context.Context, domain.ExtractedCompany) (domain.CompanySummary, error) {
			return domain.CompanySummary{}, errors.New("unexpected")
		},
	}, fakeHealth{}, nil, nil)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/companies/sync", bytes.NewBufferString(`{"name":"Attio","linkedin_url":"https://www.linkedin.com/company/attio","domain":"attio.com"}`))
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", resp.Code)
	}
}

func testCompany() domain.CompanySummary {
	now := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	return domain.CompanySummary{
		Name:          "Attio",
		LinkedInURL:   "https://www.linkedin.com/company/attio",
		Domain:        "attio.com",
		AttioRecordID: "rec_123",
		AttioWebURL:   "https://app.attio.com/rec_123",
		FirstSyncedAt: now,
		LastSyncedAt:  now,
	}
}
