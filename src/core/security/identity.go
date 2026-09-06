package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const setupSecretLifetime = 30 * time.Minute

func renewSetupSecret(state *IdentityState, now time.Time) (string, error) {
	secret, err := randomSecret(32)
	if err != nil {
		return "", err
	}
	state.SetupRequired = true
	state.SetupSecretHash = hashSecret(secret)
	state.SetupExpiresAt = now.Add(setupSecretLifetime)
	return secret, nil
}

func BootstrapOwner(setupSecret, username, password string, now time.Time) (User, error) {
	if err := ValidateUsername(username); err != nil {
		return User{}, err
	}
	passwordHash, err := HashIdentityPassword(password)
	if err != nil {
		return User{}, err
	}
	var owner User
	err = mutateIdentity(func(state *IdentityState) error {
		if !state.SetupRequired || len(state.Users) != 0 {
			return errors.New("owner setup is already complete")
		}
		if now.After(state.SetupExpiresAt) || !secretMatches(state.SetupSecretHash, setupSecret) {
			return errors.New("invalid or expired setup secret")
		}
		state.Groups[OwnerGroupID] = ownerGroup(now)
		owner = User{
			ID: uuid.NewString(), Username: strings.TrimSpace(username), Normalized: NormalizeUsername(username),
			PasswordHash: passwordHash, Enabled: true, GroupIDs: []string{OwnerGroupID}, CreatedAt: now, UpdatedAt: now,
		}
		state.Users[owner.ID] = owner
		state.SetupRequired = false
		state.SetupSecretHash = ""
		state.SetupExpiresAt = time.Time{}
		appendAudit(state, owner.ID, owner.Username, "setup.bootstrap", "user", owner.ID, now)
		return nil
	})
	return owner, err
}

func AuthenticateUser(username, password string) (User, error) {
	state := identitySnapshot()
	normalized := NormalizeUsername(username)
	var candidate User
	for _, user := range state.Users {
		if user.Normalized == normalized {
			candidate = user
			break
		}
	}
	if !verifyIdentityPasswordOrDummy(candidate.PasswordHash, password) || !candidate.Enabled {
		return User{}, errors.New("invalid credentials")
	}
	if strings.HasPrefix(candidate.PasswordHash, "$2") {
		if upgraded, err := HashIdentityPassword(password); err == nil {
			_ = mutateIdentity(func(current *IdentityState) error {
				user, ok := current.Users[candidate.ID]
				if ok && user.PasswordHash == candidate.PasswordHash {
					user.PasswordHash = upgraded
					user.UpdatedAt = time.Now()
					current.Users[user.ID] = user
				}
				return nil
			})
		}
	}
	return candidate, nil
}

func ResolvePermissions(user User, state IdentityState) map[string]bool {
	permissions := make(map[string]bool)
	for _, groupID := range user.GroupIDs {
		for _, permission := range state.Groups[groupID].Permissions {
			if IsPermission(permission) {
				permissions[permission] = true
			}
		}
	}
	return permissions
}

func SetupRequired() bool {
	return identitySnapshot().SetupRequired
}

func randomSecret(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashSecret(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func secretMatches(expected, value string) bool {
	left, leftErr := hex.DecodeString(expected)
	right, rightErr := hex.DecodeString(hashSecret(value))
	return leftErr == nil && rightErr == nil && len(left) == len(right) && subtle.ConstantTimeCompare(left, right) == 1
}
