package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type ExtractedCompany struct {
	Name            string          `json:"name"`
	LinkedInURL     string          `json:"linkedin_url"`
	WebsiteURL      string          `json:"website_url,omitempty"`
	Domain          string          `json:"domain,omitempty"`
	Description     string          `json:"description,omitempty"`
	ExtractionDebug json.RawMessage `json:"extraction_debug,omitempty"`
}

type CompanySummary struct {
	Name          string    `json:"name"`
	LinkedInURL   string    `json:"linkedin_url"`
	Domain        string    `json:"domain,omitempty"`
	AttioRecordID string    `json:"attio_record_id"`
	AttioWebURL   string    `json:"attio_web_url"`
	FirstSyncedAt time.Time `json:"first_synced_at"`
	LastSyncedAt  time.Time `json:"last_synced_at"`
}

var ErrInvalidLinkedInCompanyURL = errors.New("linkedin_url must be a LinkedIn company URL")

func NormalizeLinkedInCompanyURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidLinkedInCompanyURL
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidLinkedInCompanyURL
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "linkedin.com" && host != "www.linkedin.com" {
		return "", ErrInvalidLinkedInCompanyURL
	}

	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) < 2 || parts[0] != "company" || parts[1] == "" {
		return "", ErrInvalidLinkedInCompanyURL
	}

	slug, err := url.PathUnescape(parts[1])
	if err != nil || strings.TrimSpace(slug) == "" {
		return "", ErrInvalidLinkedInCompanyURL
	}

	return fmt.Sprintf("https://www.linkedin.com/company/%s", strings.TrimSpace(slug)), nil
}

func NormalizeDomain(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	withScheme := trimmed
	if !strings.Contains(withScheme, "://") {
		withScheme = "https://" + withScheme
	}

	parsed, err := url.Parse(withScheme)
	if err != nil || parsed.Hostname() == "" {
		return "", errors.New("domain is invalid")
	}

	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host, nil
}
