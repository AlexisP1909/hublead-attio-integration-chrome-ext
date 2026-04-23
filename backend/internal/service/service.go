package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"hublead-attio-integration/backend/internal/domain"
)

type Store interface {
	LookupCompany(ctx context.Context, normalizedLinkedInURL string, normalizedDomain string) (domain.CompanySummary, bool, error)
	UpsertCompany(ctx context.Context, input UpsertCompanyInput) (domain.CompanySummary, error)
}

type AttioClient interface {
	AssertCompany(ctx context.Context, input AttioCompanyInput) (AttioCompanyResult, error)
}

type Service struct {
	store Store
	attio AttioClient
}

type UpsertCompanyInput struct {
	NormalizedLinkedInURL string
	NormalizedDomain      string
	CompanyName           string
	AttioRecordID         string
	AttioWebURL           string
	RawExtractedJSON      json.RawMessage
}

type AttioCompanyInput struct {
	Name        string
	LinkedInURL string
	Domain      string
	Description string
}

type AttioCompanyResult struct {
	RecordID string
	WebURL   string
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

var ErrNotFound = errors.New("company not found")

func New(store Store, attio AttioClient) *Service {
	return &Service{store: store, attio: attio}
}

func (s *Service) Lookup(ctx context.Context, linkedinURL string, rawDomain string) (domain.CompanySummary, error) {
	normalizedLinkedInURL, err := domain.NormalizeLinkedInCompanyURL(linkedinURL)
	if err != nil {
		return domain.CompanySummary{}, ValidationError{Message: "linkedin_url must be a LinkedIn company URL"}
	}

	normalizedDomain, err := domain.NormalizeDomain(rawDomain)
	if err != nil {
		return domain.CompanySummary{}, ValidationError{Message: "domain is invalid"}
	}

	company, found, err := s.store.LookupCompany(ctx, normalizedLinkedInURL, normalizedDomain)
	if err != nil {
		return domain.CompanySummary{}, err
	}
	if !found {
		return domain.CompanySummary{}, ErrNotFound
	}
	return company, nil
}

func (s *Service) Sync(ctx context.Context, input domain.ExtractedCompany) (domain.CompanySummary, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.CompanySummary{}, ValidationError{Message: "name is required"}
	}

	normalizedLinkedInURL, err := domain.NormalizeLinkedInCompanyURL(input.LinkedInURL)
	if err != nil {
		return domain.CompanySummary{}, ValidationError{Message: "linkedin_url must be a LinkedIn company URL"}
	}

	normalizedDomain, err := domain.NormalizeDomain(input.Domain)
	if err != nil {
		return domain.CompanySummary{}, ValidationError{Message: "domain is invalid"}
	}
	if normalizedDomain == "" {
		return domain.CompanySummary{}, ValidationError{Message: "domain is required"}
	}

	attioCompany, err := s.attio.AssertCompany(ctx, AttioCompanyInput{
		Name:        name,
		LinkedInURL: normalizedLinkedInURL,
		Domain:      normalizedDomain,
		Description: strings.TrimSpace(input.Description),
	})
	if err != nil {
		return domain.CompanySummary{}, err
	}

	rawJSON, err := json.Marshal(input)
	if err != nil {
		return domain.CompanySummary{}, err
	}

	return s.store.UpsertCompany(ctx, UpsertCompanyInput{
		NormalizedLinkedInURL: normalizedLinkedInURL,
		NormalizedDomain:      normalizedDomain,
		CompanyName:           name,
		AttioRecordID:         attioCompany.RecordID,
		AttioWebURL:           attioCompany.WebURL,
		RawExtractedJSON:      rawJSON,
	})
}
