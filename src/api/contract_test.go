package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/security"
)

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
