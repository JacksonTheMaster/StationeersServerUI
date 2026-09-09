package security

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const developmentPasswordHash = "$2a$10$7QQhPkNAfT.MXhJhnnodXOyn3KKE/1eu7nYb0y2O1UBoAWc0Y/fda" // admin

var ErrRecoveryOwnerStillRequired = errors.New("recovery account is still the only owner")

func RecoverOwner(username, password string, now time.Time) (User, error) {
	if err := ValidateUsername(username); err != nil {
		return User{}, err
	}
	hash, err := HashIdentityPassword(password)
	if err != nil {
		return User{}, err
	}
	return recoverOwnerWithHash(username, hash, "local-cli", now)
}

func EnableDevelopmentOwner(now time.Time) (User, error) {
	return recoverOwnerWithHash("admin", developmentPasswordHash, "development", now)
}

func RemoveRecoveryOwner(now time.Time) (bool, error) {
	state := identitySnapshot()
	if _, ok := recoveryUser(state); !ok {
		return false, nil
	}

	removed := false
	err := mutateIdentity(func(state *IdentityState) error {
		recovery, ok := recoveryUser(*state)
		if !ok {
			return nil
		}

		otherOwner := false
		for id, user := range state.Users {
			if id != recovery.ID && user.Enabled && contains(user.GroupIDs, OwnerGroupID) {
				otherOwner = true
				break
			}
		}
		if !otherOwner && len(state.Users) > 1 {
			return ErrRecoveryOwnerStillRequired
		}

		delete(state.Users, recovery.ID)
		for id, session := range state.Sessions {
			if session.UserID == recovery.ID {
				delete(state.Sessions, id)
			}
		}
		for id, token := range state.Tokens {
			if token.OwnerID == recovery.ID {
				delete(state.Tokens, id)
			}
		}
		if len(state.Users) == 0 {
			state.SetupRequired = true
		}
		appendAudit(state, "", "local-cli", "owner.recovery.remove", "user", recovery.ID, now)
		removed = true
		return nil
	})
	return removed, err
}

func recoveryUser(state IdentityState) (User, bool) {
	for _, user := range state.Users {
		if user.Normalized != "recovery" {
			continue
		}
		for i := len(state.Audit) - 1; i >= 0; i-- {
			event := state.Audit[i]
			if event.TargetID == user.ID && event.Action == "owner.recover" {
				return user, true
			}
		}
	}
	return User{}, false
}

func recoverOwnerWithHash(username, hash, actor string, now time.Time) (User, error) {
	var recovered User
	err := mutateIdentity(func(state *IdentityState) error {
		if _, ok := state.Groups[OwnerGroupID]; !ok {
			state.Groups[OwnerGroupID] = ownerGroup(now)
		}
		normalized := NormalizeUsername(username)
		for id, user := range state.Users {
			if user.Normalized != normalized {
				continue
			}
			user.Username = strings.TrimSpace(username)
			user.PasswordHash = hash
			user.Enabled = true
			if !contains(user.GroupIDs, OwnerGroupID) {
				user.GroupIDs = append(user.GroupIDs, OwnerGroupID)
			}
			user.UpdatedAt = now
			state.Users[id] = user
			recovered = user
			break
		}
		if recovered.ID == "" {
			recovered = User{
				ID: uuid.NewString(), Username: strings.TrimSpace(username), Normalized: normalized,
				PasswordHash: hash, Enabled: true, GroupIDs: []string{OwnerGroupID}, CreatedAt: now, UpdatedAt: now,
			}
			state.Users[recovered.ID] = recovered
		}
		state.Sessions = make(map[string]Session)
		state.Tokens = make(map[string]Token)
		state.SetupRequired = false
		appendAudit(state, "", actor, "owner.recover", "user", recovered.ID, now)
		return nil
	})
	return recovered, err
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
