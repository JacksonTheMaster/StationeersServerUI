package security

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func RecoverOwner(username, password string, now time.Time) (User, error) {
	if err := ValidateUsername(username); err != nil {
		return User{}, err
	}
	hash, err := HashIdentityPassword(password)
	if err != nil {
		return User{}, err
	}
	var recovered User
	err = mutateIdentity(func(state *IdentityState) error {
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
		appendAudit(state, "", "local-cli", "owner.recover", "user", recovered.ID, now)
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
