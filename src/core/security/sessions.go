package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	SessionCookieName       = "SSUISession"
	SessionIdleLifetime     = 24 * time.Hour
	SessionAbsoluteLifetime = 30 * 24 * time.Hour
	sessionTouchInterval    = 5 * time.Minute
)

type SessionCredential struct {
	Value string
	CSRF  string
}

func CreateSession(userID string, now time.Time) (Session, SessionCredential, error) {
	secret, err := randomSecret(32)
	if err != nil {
		return Session{}, SessionCredential{}, err
	}
	csrf := deriveCSRF(secret)
	session := Session{
		ID: uuid.NewString(), SecretHash: hashSecret(secret), CSRFHash: hashSecret(csrf), UserID: userID,
		CreatedAt: now, LastUsedAt: now, IdleExpiresAt: now.Add(SessionIdleLifetime), AbsoluteExpiresAt: now.Add(SessionAbsoluteLifetime),
	}
	err = mutateIdentity(func(state *IdentityState) error {
		user, ok := state.Users[userID]
		if !ok || !user.Enabled {
			return errors.New("user is disabled or missing")
		}
		state.Sessions[session.ID] = session
		return nil
	})
	return session, SessionCredential{Value: session.ID + "." + secret, CSRF: csrf}, err
}

func AuthenticateSession(value string, now time.Time) (Principal, Session, error) {
	id, secret, ok := strings.Cut(value, ".")
	if !ok || id == "" || secret == "" {
		return Principal{}, Session{}, errors.New("invalid session")
	}
	state := identitySnapshot()
	session, ok := state.Sessions[id]
	if !ok || !secretMatches(session.SecretHash, secret) {
		return Principal{}, Session{}, errors.New("invalid session")
	}
	if !session.AbsoluteExpiresAt.After(now) || !session.IdleExpiresAt.After(now) {
		_ = RevokeSession(id)
		return Principal{}, Session{}, errors.New("session expired")
	}
	user, ok := state.Users[session.UserID]
	if !ok || !user.Enabled {
		_ = RevokeSession(id)
		return Principal{}, Session{}, errors.New("user is disabled or missing")
	}
	if now.Sub(session.LastUsedAt) >= sessionTouchInterval {
		_ = mutateIdentity(func(current *IdentityState) error {
			stored, exists := current.Sessions[id]
			if exists && stored.SecretHash == session.SecretHash {
				stored.LastUsedAt = now
				stored.IdleExpiresAt = now.Add(SessionIdleLifetime)
				if stored.IdleExpiresAt.After(stored.AbsoluteExpiresAt) {
					stored.IdleExpiresAt = stored.AbsoluteExpiresAt
				}
				current.Sessions[id] = stored
			}
			return nil
		})
	}
	return principalForUser(user, state, id, "session"), session, nil
}

func ValidateSessionCSRF(session Session, value string) bool {
	return value != "" && secretMatches(session.CSRFHash, value)
}

func CSRFForSessionCredential(value string) string {
	_, secret, ok := strings.Cut(value, ".")
	if !ok {
		return ""
	}
	return deriveCSRF(secret)
}

func deriveCSRF(secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("ssui-csrf-v1"))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func RevokeSession(id string) error {
	return revokeSession(id, Principal{}, time.Now())
}

func RevokeSessionAs(id string, actor Principal, now time.Time) error {
	return revokeSession(id, actor, now)
}

func revokeSession(id string, actor Principal, now time.Time) error {
	return mutateIdentity(func(state *IdentityState) error {
		if _, ok := state.Sessions[id]; !ok {
			return errors.New("session not found")
		}
		delete(state.Sessions, id)
		if actor.Username != "" {
			appendAudit(state, actor.UserID, actor.Username, "session.revoke", "session", id, now)
		}
		return nil
	})
}

func revokeUserCredentials(state *IdentityState, userID string) {
	for id, session := range state.Sessions {
		if session.UserID == userID {
			delete(state.Sessions, id)
		}
	}
	for id, token := range state.Tokens {
		if token.OwnerID == userID {
			delete(state.Tokens, id)
		}
	}
}

func principalForUser(user User, state IdentityState, credentialID, credential string) Principal {
	return Principal{
		UserID: user.ID, Username: user.Username, CredentialID: credentialID,
		Credential: credential, Permissions: ResolvePermissions(user, state),
	}
}
