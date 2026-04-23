package attio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"hublead-attio-integration/backend/internal/service"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type AuthError struct {
	StatusCode int
	Message    string
}

func (e AuthError) Error() string {
	return e.Message
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	return e.Message
}

func NewClient(baseURL string, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    httpClient,
	}
}

func (c *Client) AssertCompany(ctx context.Context, input service.AttioCompanyInput) (service.AttioCompanyResult, error) {
	values := map[string]any{
		"domains":  []string{input.Domain},
		"name":     input.Name,
		"linkedin": input.LinkedInURL,
	}
	if input.Description != "" {
		values["description"] = input.Description
	}

	body, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"values": values,
		},
	})
	if err != nil {
		return service.AttioCompanyResult{}, err
	}

	endpoint := c.baseURL + "/v2/objects/companies/records?matching_attribute=domains"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return service.AttioCompanyResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return service.AttioCompanyResult{}, APIError{StatusCode: http.StatusBadGateway, Message: "Attio request failed"}
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return service.AttioCompanyResult{}, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return service.AttioCompanyResult{}, AuthError{StatusCode: resp.StatusCode, Message: "Attio credentials were rejected"}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return service.AttioCompanyResult{}, APIError{StatusCode: resp.StatusCode, Message: attioErrorMessage(payload, resp.StatusCode)}
	}

	var parsed struct {
		Data struct {
			ID struct {
				RecordID string `json:"record_id"`
			} `json:"id"`
			WebURL string `json:"web_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return service.AttioCompanyResult{}, APIError{StatusCode: http.StatusBadGateway, Message: "Attio returned invalid JSON"}
	}

	if parsed.Data.ID.RecordID == "" || parsed.Data.WebURL == "" {
		return service.AttioCompanyResult{}, APIError{StatusCode: http.StatusBadGateway, Message: "Attio response missing record details"}
	}

	return service.AttioCompanyResult{RecordID: parsed.Data.ID.RecordID, WebURL: parsed.Data.WebURL}, nil
}

func attioErrorMessage(payload []byte, status int) string {
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err == nil {
		for _, key := range []string{"error", "message"} {
			if value, ok := body[key].(string); ok && value != "" {
				return value
			}
		}
	}
	if len(payload) > 0 {
		return fmt.Sprintf("Attio returned HTTP %d", status)
	}
	return "Attio request failed"
}

func IsAuthError(err error) bool {
	var authErr AuthError
	return errors.As(err, &authErr)
}

func IsAPIError(err error) bool {
	var apiErr APIError
	return errors.As(err, &apiErr)
}
