package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/security"
)

type credentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SetupStatusHandler(w http.ResponseWriter, r *http.Request) {
	WriteData(w, http.StatusOK, map[string]bool{"setupRequired": security.SetupRequired()})
}

func BootstrapOwnerHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SetupSecret string `json:"setupSecret"`
		Username    string `json:"username"`
		Password    string `json:"password"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	owner, err := security.BootstrapOwner(request.SetupSecret, request.Username, request.Password, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "setup_failed", err.Error())
		return
	}
	respondWithNewSession(w, owner)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if security.SetupRequired() {
		WriteError(w, http.StatusServiceUnavailable, "setup_required", "Owner setup is required")
		return
	}
	var request credentialsRequest
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	user, ok := authenticateLogin(w, r, request.Username, request.Password)
	if !ok {
		return
	}
	respondWithNewSession(w, user)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		WriteError(w, http.StatusForbidden, "origin_denied", "Origin is not allowed")
		return
	}
	if cookie, err := r.Cookie(security.SessionCookieName); err == nil {
		_, session, authErr := security.AuthenticateSession(cookie.Value, time.Now())
		if authErr == nil {
			if !security.ValidateSessionCSRF(session, r.Header.Get("X-SSUI-CSRF")) {
				WriteError(w, http.StatusForbidden, "csrf_failed", "CSRF validation failed")
				return
			}
			_ = security.RevokeSession(session.ID)
		}
	}
	clearSessionCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func SessionInfoHandler(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	user, ok := security.GetUser(principal.UserID)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User no longer exists")
		return
	}
	response := sessionResponse(user, principal)
	if session, ok := SessionFromContext(r.Context()); ok {
		response["expiresAt"] = session.AbsoluteExpiresAt
	}
	WriteData(w, http.StatusOK, response)
}

func ChangeOwnPasswordHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	if principal.Credential != "session" {
		WriteError(w, http.StatusForbidden, "session_required", "Password changes require a browser session")
		return
	}
	var request struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := security.ChangeOwnPassword(principal, request.CurrentPassword, request.NewPassword, time.Now()); err != nil {
		WriteError(w, http.StatusBadRequest, "password_failed", err.Error())
		return
	}
	clearSessionCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	WriteData(w, http.StatusOK, map[string]any{"users": security.ListUsers()})
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		GroupIDs []string `json:"groupIds"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	user, err := security.CreateUser(principal, request.Username, request.Password, request.GroupIDs, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "user_failed", err.Error())
		return
	}
	WriteData(w, http.StatusCreated, user)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username *string   `json:"username"`
		Password *string   `json:"password"`
		Enabled  *bool     `json:"enabled"`
		GroupIDs *[]string `json:"groupIds"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	user, err := security.UpdateUser(principal, r.PathValue("id"), security.UserUpdate{
		Username: request.Username, Password: request.Password, Enabled: request.Enabled, GroupIDs: request.GroupIDs,
	}, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "user_failed", err.Error())
		return
	}
	WriteData(w, http.StatusOK, user)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	if err := security.DeleteUser(principal, r.PathValue("id"), time.Now()); err != nil {
		WriteError(w, http.StatusBadRequest, "user_failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GroupsHandler(w http.ResponseWriter, r *http.Request) {
	WriteData(w, http.StatusOK, map[string]any{
		"groups": security.ListGroups(), "permissions": security.AllPermissions,
	})
}

func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	group, err := security.CreateGroup(principal, request.Name, request.Description, request.Permissions, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "group_failed", err.Error())
		return
	}
	WriteData(w, http.StatusCreated, group)
}

func UpdateGroupHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	group, err := security.UpdateGroup(principal, r.PathValue("id"), request.Name, request.Description, request.Permissions, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "group_failed", err.Error())
		return
	}
	WriteData(w, http.StatusOK, group)
}

func DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	if err := security.DeleteGroup(principal, r.PathValue("id"), time.Now()); err != nil {
		WriteError(w, http.StatusBadRequest, "group_failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func TokensHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	tokens := security.ListTokens(principal)
	items := make([]tokenResponse, 0, len(tokens))
	for _, token := range tokens {
		items = append(items, publicToken(token))
	}
	WriteData(w, http.StatusOK, map[string]any{"tokens": items})
}

func CreateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name      string     `json:"name"`
		Scopes    []string   `json:"scopes"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := DecodeJSON(w, r, &request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	token, secret, err := security.CreateToken(principal.UserID, request.Name, request.Scopes, request.ExpiresAt, time.Now())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "token_failed", err.Error())
		return
	}
	WriteData(w, http.StatusCreated, map[string]any{"token": publicToken(token), "secret": secret})
}

func DeleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	token, ok := security.GetToken(r.PathValue("id"))
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "Token not found")
		return
	}
	if token.OwnerID != principal.UserID && !principal.Permissions[security.PermissionSecurityManage] {
		WriteError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	if err := security.RevokeToken(token.ID, principal, time.Now()); err != nil {
		WriteError(w, http.StatusInternalServerError, "save_failed", "Could not revoke token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func SessionsHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	sessions := security.ListSessions(principal)
	items := make([]sessionListResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, sessionListResponse{
			ID: session.ID, UserID: session.UserID, CreatedAt: session.CreatedAt,
			LastUsedAt: session.LastUsedAt, IdleExpiresAt: session.IdleExpiresAt,
			AbsoluteExpiresAt: session.AbsoluteExpiresAt,
		})
	}
	WriteData(w, http.StatusOK, map[string]any{"sessions": items})
}

type tokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	OwnerID    string     `json:"ownerId"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

func publicToken(token security.Token) tokenResponse {
	return tokenResponse{
		ID: token.ID, Name: token.Name, OwnerID: token.OwnerID, Scopes: token.Scopes,
		CreatedAt: token.CreatedAt, ExpiresAt: token.ExpiresAt,
		LastUsedAt: token.LastUsedAt, RevokedAt: token.RevokedAt,
	}
}

type sessionListResponse struct {
	ID                string    `json:"id"`
	UserID            string    `json:"userId"`
	CreatedAt         time.Time `json:"createdAt"`
	LastUsedAt        time.Time `json:"lastUsedAt"`
	IdleExpiresAt     time.Time `json:"idleExpiresAt"`
	AbsoluteExpiresAt time.Time `json:"absoluteExpiresAt"`
}

func DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	session, ok := security.GetSession(r.PathValue("id"))
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "Session not found")
		return
	}
	if session.UserID != principal.UserID && !principal.Permissions[security.PermissionSecurityManage] {
		WriteError(w, http.StatusForbidden, "forbidden", "Permission denied")
		return
	}
	if err := security.RevokeSessionAs(session.ID, principal, time.Now()); err != nil {
		WriteError(w, http.StatusInternalServerError, "save_failed", "Could not revoke session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func AuditHandler(w http.ResponseWriter, r *http.Request) {
	WriteData(w, http.StatusOK, map[string]any{"events": security.ListAudit()})
}

func CapabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	permissions := make([]string, 0, len(principal.Permissions))
	for permission := range principal.Permissions {
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)
	WriteData(w, http.StatusOK, map[string]any{
		"apiVersion": "v3",
		"backend": map[string]any{
			"version": config.GetVersion(),
		},
		"permissions": permissions,
	})
}

func respondWithNewSession(w http.ResponseWriter, user security.User) {
	now := time.Now()
	session, credential, err := security.CreateSession(user.ID, now)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "session_failed", "Could not create session")
		return
	}
	setSessionCookies(w, session, credential)
	principal, _, err := security.AuthenticateSession(credential.Value, now)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "session_failed", "Could not read new session")
		return
	}
	response := sessionResponse(user, principal)
	response["csrf"] = credential.CSRF
	response["expiresAt"] = session.AbsoluteExpiresAt
	WriteData(w, http.StatusOK, response)
}

func sessionResponse(user security.User, principal security.Principal) map[string]any {
	permissions := make([]string, 0, len(principal.Permissions))
	for permission := range principal.Permissions {
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)
	return map[string]any{
		"user": map[string]any{
			"id": user.ID, "username": user.Username, "groupIds": user.GroupIDs,
		},
		"credentialType": principal.Credential,
		"credentialId":   principal.CredentialID,
		"permissions":    permissions,
	}
}

func setSessionCookies(w http.ResponseWriter, session security.Session, credential security.SessionCredential) {
	maxAge := int(time.Until(session.AbsoluteExpiresAt).Seconds())
	http.SetCookie(w, &http.Cookie{
		Name: security.SessionCookieName, Value: credential.Value, Path: "/", Expires: session.AbsoluteExpiresAt,
		MaxAge: maxAge, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name: security.CSRFCookieName, Value: credential.CSRF, Path: "/", Expires: session.AbsoluteExpiresAt,
		MaxAge: maxAge, Secure: true, SameSite: http.SameSiteStrictMode,
	})
}

func clearSessionCookies(w http.ResponseWriter) {
	for _, name := range []string{security.SessionCookieName, security.CSRFCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name: name, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0),
			HttpOnly: name == security.SessionCookieName, Secure: true, SameSite: http.SameSiteStrictMode,
		})
	}
}
