package domain

import "testing"

func TestNormalizeLinkedInCompanyURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "www company page", raw: "https://www.linkedin.com/company/attio/?view=all", want: "https://www.linkedin.com/company/attio"},
		{name: "bare linkedin host", raw: "https://linkedin.com/company/hublead", want: "https://www.linkedin.com/company/hublead"},
		{name: "profile page rejected", raw: "https://www.linkedin.com/in/person", wantErr: true},
		{name: "non linkedin rejected", raw: "https://example.com/company/attio", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeLinkedInCompanyURL(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "https://www.attio.com/", want: "attio.com"},
		{raw: "WWW.Example.COM/path", want: "example.com"},
		{raw: "hublead.io", want: "hublead.io"},
		{raw: "", want: ""},
	}

	for _, tt := range tests {
		got, err := NormalizeDomain(tt.raw)
		if err != nil {
			t.Fatalf("NormalizeDomain(%q) error: %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("NormalizeDomain(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}
