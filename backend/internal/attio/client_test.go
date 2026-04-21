package attio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hublead-attio-integration/backend/internal/service"
)

func TestAssertCompanySendsExpectedRequestAndParsesResponse(t *testing.T) {
	var captured struct {
		Method        string
		Path          string
		Query         string
		Authorization string
		Body          map[string]any
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Method = r.Method
		captured.Path = r.URL.Path
		captured.Query = r.URL.RawQuery
		captured.Authorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&captured.Body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":      map[string]string{"record_id": "rec_123"},
				"web_url": "https://app.attio.com/rec_123",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret-token", server.Client())
	result, err := client.AssertCompany(t.Context(), service.AttioCompanyInput{
		Name:        "Attio",
		LinkedInURL: "https://www.linkedin.com/company/attio",
		Domain:      "attio.com",
		Description: "Customer relationship magic.",
	})
	if err != nil {
		t.Fatalf("AssertCompany error: %v", err)
	}

	if result.RecordID != "rec_123" || result.WebURL != "https://app.attio.com/rec_123" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if captured.Method != http.MethodPut {
		t.Fatalf("method = %s", captured.Method)
	}
	if captured.Path != "/v2/objects/companies/records" {
		t.Fatalf("path = %s", captured.Path)
	}
	if captured.Query != "matching_attribute=domains" {
		t.Fatalf("query = %s", captured.Query)
	}
	if captured.Authorization != "Bearer secret-token" {
		t.Fatalf("authorization = %s", captured.Authorization)
	}

	values := captured.Body["data"].(map[string]any)["values"].(map[string]any)
	if values["name"] != "Attio" || values["linkedin"] != "https://www.linkedin.com/company/attio" {
		t.Fatalf("unexpected values: %#v", values)
	}
	domains := values["domains"].([]any)
	if len(domains) != 1 || domains[0] != "attio.com" {
		t.Fatalf("unexpected domains: %#v", domains)
	}
}

func TestAssertCompanyMapsAttioFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"bad upstream"}`, http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret-token", server.Client())
	_, err := client.AssertCompany(t.Context(), service.AttioCompanyInput{Name: "Attio", LinkedInURL: "https://www.linkedin.com/company/attio", Domain: "attio.com"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAPIError(err) {
		t.Fatalf("expected APIError, got %T", err)
	}
}
