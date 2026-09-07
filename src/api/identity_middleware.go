package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/core/security"
)

type identityContextKey string

const (
	principalContextKey identityContextKey = "principal"
	sessionContextKey   identityContextKey = "session"
)

func IdentityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if security.SetupRequired() {
			WriteError(w, http.StatusServiceUnavailable, "setup_required", "Owner setup is required")
			return
		}
		principal, session, err := authenticateRequest(r)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}
		if principal.Credential == "session" && changesState(r.Method) {
			if !sameOrigin(r) || !security.ValidateSessionCSRF(session, r.Header.Get("X-SSUI-CSRF")) {
				WriteError(w, http.StatusForbidden, "csrf_failed", "CSRF validation failed")
				return
			}
		}
		ctx := context.WithValue(r.Context(), principalContextKey, principal)
		if principal.Credential == "session" {
			ctx = context.WithValue(ctx, sessionContextKey, session)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Require(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}
		if !principal.Permissions[permission] {
			WriteError(w, http.StatusForbidden, "forbidden", "Permission denied")
			return
		}
		next(w, r)
	}
}

func PrincipalFromContext(ctx context.Context) (security.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(security.Principal)
	return principal, ok
}

func SessionFromContext(ctx context.Context) (security.Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(security.Session)
	return session, ok
}

func PageIdentityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if security.SetupRequired() {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}
		principal, session, err := authenticateRequest(r)
		if err != nil || principal.Credential != "session" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey, principal)
		ctx = context.WithValue(ctx, sessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authenticateRequest(r *http.Request) (security.Principal, security.Session, error) {
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		principal, _, err := security.AuthenticateToken(strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")), time.Now())
		return principal, security.Session{}, err
	}
	cookie, err := r.Cookie(security.SessionCookieName)
	if err != nil {
		return security.Principal{}, security.Session{}, err
	}
	return security.AuthenticateSession(cookie.Value, time.Now())
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return origin == scheme+"://"+r.Host
}

func changesState(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}
