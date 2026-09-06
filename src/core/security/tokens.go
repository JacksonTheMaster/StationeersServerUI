package security

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const tokenPrefix = "ssui_pat_"

func CreateToken(ownerID, name string, scopes []string, expiresAt *time.Time, now time.Time) (Token, string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 80 {
		return Token{}, "", errors.New("token name must be between 1 and 80 characters")
	}
	if expiresAt != nil && !expiresAt.After(now) {
		return Token{}, "", errors.New("token expiry must be in the future")
	}
	secret, err := randomSecret(32)
	if err != nil {
		return Token{}, "", err
	}
	token := Token{
		ID: uuid.NewString(), Name: name, SecretHash: hashSecret(secret), OwnerID: ownerID,
		Scopes: append([]string(nil), scopes...), CreatedAt: now, ExpiresAt: expiresAt,
	}
	var owner User
	err = mutateIdentity(func(current *IdentityState) error {
		var ok bool
		owner, ok = current.Users[ownerID]
		if !ok || !owner.Enabled {
			return errors.New("token owner is disabled or missing")
		}
		permissions := ResolvePermissions(owner, *current)
		granted := make([]string, 0, len(scopes))
		seen := make(map[string]bool)
		for _, scope := range scopes {
			if !IsPermission(scope) || !permissions[scope] {
				return errors.New("token scope exceeds owner permissions")
			}
			if !seen[scope] {
				seen[scope] = true
				granted = append(granted, scope)
			}
		}
		if len(granted) == 0 {
			return errors.New("token requires at least one scope")
		}
		token.Scopes = granted
		current.Tokens[token.ID] = token
		appendAudit(current, owner.ID, owner.Username, "token.create", "token", token.ID, now)
		return nil
	})
	return token, tokenPrefix + token.ID + "_" + secret, err
}

func AuthenticateToken(value string, now time.Time) (Principal, Token, error) {
	if !strings.HasPrefix(value, tokenPrefix) {
		return Principal{}, Token{}, errors.New("invalid token")
	}
	id, secret, ok := strings.Cut(strings.TrimPrefix(value, tokenPrefix), "_")
	if !ok || id == "" || secret == "" {
		return Principal{}, Token{}, errors.New("invalid token")
	}
	state := identitySnapshot()
	token, ok := state.Tokens[id]
	if !ok || token.RevokedAt != nil || !secretMatches(token.SecretHash, secret) {
		return Principal{}, Token{}, errors.New("invalid token")
	}
	if token.ExpiresAt != nil && !token.ExpiresAt.After(now) {
		return Principal{}, Token{}, errors.New("token expired")
	}
	user, ok := state.Users[token.OwnerID]
	if !ok || !user.Enabled {
		return Principal{}, Token{}, errors.New("token owner is disabled or missing")
	}
	ownerPermissions := ResolvePermissions(user, state)
	granted := make(map[string]bool)
	for _, scope := range token.Scopes {
		if ownerPermissions[scope] {
			granted[scope] = true
		}
	}
	if token.LastUsedAt == nil || now.Sub(*token.LastUsedAt) >= sessionTouchInterval {
		_ = mutateIdentity(func(current *IdentityState) error {
			stored, exists := current.Tokens[id]
			if exists && stored.SecretHash == token.SecretHash && stored.RevokedAt == nil {
				stored.LastUsedAt = &now
				current.Tokens[id] = stored
			}
			return nil
		})
	}
	return Principal{
		UserID: user.ID, Username: user.Username, CredentialID: token.ID,
		Credential: "token", Permissions: granted,
	}, token, nil
}

func RevokeToken(id string, actor Principal, now time.Time) error {
	return mutateIdentity(func(state *IdentityState) error {
		token, ok := state.Tokens[id]
		if !ok {
			return errors.New("token not found")
		}
		token.RevokedAt = &now
		state.Tokens[id] = token
		appendAudit(state, actor.UserID, actor.Username, "token.revoke", "token", id, now)
		return nil
	})
}
