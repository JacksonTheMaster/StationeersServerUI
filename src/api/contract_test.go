package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/security"
)

func TestJSONBoundaryNormalizesLegacyResponses(t *testing.T) {
	handler := JSONBoundary(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"isRunning": true})
	})
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodGet, "/api/v3/server/status", nil))

	var body Envelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	data, ok := body.Data.(map[string]any)
	if !ok || data["isRunning"] != true || body.Error != nil {
		t.Fatalf("unexpected v3 response: %#v", body)
	}
}

func TestJSONBoundaryNormalizesErrors(t *testing.T) {
	handler := JSONBoundary(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not today", http.StatusConflict)
	})
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodPost, "/api/v3/test", nil))

	var body Envelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusConflict || body.Error == nil || body.Error.Message != "not today" || body.Data != nil {
		t.Fatalf("unexpected v3 error: %#v", body)
	}
}

func TestIdentityMiddlewareRequiresCSRFForSessions(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(workingDirectory)

	now := time.Now()
	secret, err := security.InitializeIdentity(nil, now)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := security.BootstrapOwner(secret, "admin", "correct horse battery staple", now)
	if err != nil {
		t.Fatal(err)
	}
	_, credential, err := security.CreateSession(owner.ID, now)
	if err != nil {
		t.Fatal(err)
	}

	handler := IdentityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := PrincipalFromContext(r.Context()); !ok {
			t.Error("principal missing from request")
		}
		WriteData(w, http.StatusOK, map[string]bool{"ok": true})
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v3/server/start", bytes.NewReader(nil))
	request.AddCookie(&http.Cookie{Name: security.SessionCookieName, Value: credential.Value})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("POST without CSRF returned %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v3/server/start", bytes.NewReader(nil))
	request.AddCookie(&http.Cookie{Name: security.SessionCookieName, Value: credential.Value})
	request.Header.Set("X-SSUI-CSRF", credential.CSRF)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("POST with CSRF returned %d: %s", response.Code, response.Body.String())
	}
}
