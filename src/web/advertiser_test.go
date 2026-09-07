package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
)

func TestNormalizeAdvertiserOverride(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		value     string
		want      string
		wantError bool
	}{
		{name: "disabled", mode: "disabled", want: ""},
		{name: "automatic", mode: "auto", want: "auto"},
		{name: "ipv4", mode: "ipv4", value: " 203.0.113.10 ", want: "203.0.113.10"},
		{name: "dns", mode: "dns", value: "Server.Example.COM", want: "server.example.com"},
		{name: "ipv6 rejected", mode: "ipv4", value: "2001:db8::1", wantError: true},
		{name: "invalid dns", mode: "dns", value: "not a host!", wantError: true},
		{name: "unknown mode", mode: "other", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeAdvertiserOverride(test.mode, test.value)
			if test.wantError && err == nil {
				t.Fatal("expected validation error")
			}
			if !test.wantError && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestAdvertiserOverrideRejectsWrongMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v3/advertiser/override", nil)
	response := httptest.NewRecorder()
	SaveAdvertiserOverrideHandler(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
}

func TestConfigTemplateParses(t *testing.T) {
	if _, err := template.ParseFiles("../../SSUI/onboard_bundled/ui/config.html"); err != nil {
		t.Fatalf("config template failed to parse: %v", err)
	}
}
