package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/google/uuid"
)

const maxIdentityFileSize = 16 << 20
const maxAuditEvents = 1000

var identityMu sync.RWMutex
var identityState IdentityState
var identityFile string

func InitializeIdentity(legacyUsers map[string]string, now time.Time) (string, error) {
	identityMu.Lock()
	defer identityMu.Unlock()
	identityFile = filepath.Join(config.GetSSUIFolder(), "security", "identity.json")

	state, err := readIdentity(identityFile)
	if err == nil {
		identityState = state
		if !state.SetupRequired || state.SetupExpiresAt.After(now) {
			return "", nil
		}
		secret, err := renewSetupSecret(&identityState, now)
		if err != nil {
			return "", err
		}
		return secret, saveIdentityLocked()
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	identityState = newIdentityState()
	if importLegacyUsers(&identityState, legacyUsers, now) {
		identityState.SetupRequired = false
		appendAudit(&identityState, "", "migration", "identity.migrate", "users", "", now)
		return "", saveIdentityLocked()
	}
	secret, err := renewSetupSecret(&identityState, now)
	if err != nil {
		return "", err
	}
	return secret, saveIdentityLocked()
}

func newIdentityState() IdentityState {
	return IdentityState{
		SchemaVersion: IdentitySchemaVersion,
		SetupRequired: true,
		Users:         make(map[string]User),
		Groups:        make(map[string]Group),
		Sessions:      make(map[string]Session),
		Tokens:        make(map[string]Token),
		Audit:         make([]AuditEvent, 0),
	}
}

func importLegacyUsers(state *IdentityState, users map[string]string, now time.Time) bool {
	usernames := make([]string, 0, len(users))
	for username := range users {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	groupAdded := false
	seen := make(map[string]bool)
	for _, username := range usernames {
		passwordHash := users[username]
		normalized := NormalizeUsername(username)
		if strings.HasPrefix(strings.ToLower(username), "apikey-") || strings.TrimSpace(username) == "" || passwordHash == "" {
			continue
		}
		if seen[normalized] || ValidateUsername(username) != nil {
			continue
		}
		seen[normalized] = true
		if !groupAdded {
			state.Groups[OwnerGroupID] = ownerGroup(now)
			groupAdded = true
		}
		id := uuid.NewString()
		state.Users[id] = User{
			ID:           id,
			Username:     strings.TrimSpace(username),
			Normalized:   normalized,
			PasswordHash: passwordHash,
			Enabled:      true,
			GroupIDs:     []string{OwnerGroupID},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}
	return groupAdded
}

func ownerGroup(now time.Time) Group {
	return Group{
		ID:          OwnerGroupID,
		Name:        "Owner",
		Description: "Full control of this SSUI backend",
		System:      true,
		Permissions: append([]string(nil), AllPermissions...),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func readIdentity(path string) (IdentityState, error) {
	var state IdentityState
	file, err := os.Open(path)
	if err != nil {
		return state, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxIdentityFileSize+1))
	if err != nil {
		return state, err
	}
	if len(data) > maxIdentityFileSize {
		return state, fmt.Errorf("identity file exceeds %d bytes", maxIdentityFileSize)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("decode identity file: %w", err)
	}
	if state.SchemaVersion != IdentitySchemaVersion {
		return state, fmt.Errorf("unsupported identity schema %d", state.SchemaVersion)
	}
	if state.Users == nil || state.Groups == nil || state.Sessions == nil || state.Tokens == nil {
		return state, fmt.Errorf("identity file is missing required collections")
	}
	if state.Audit == nil {
		state.Audit = make([]AuditEvent, 0)
	}
	if err := validateIdentity(state); err != nil {
		return state, err
	}
	return state, nil
}

func mutateIdentity(change func(*IdentityState) error) error {
	identityMu.Lock()
	defer identityMu.Unlock()
	next := cloneIdentity(identityState)
	if err := change(&next); err != nil {
		return err
	}
	if err := validateIdentity(next); err != nil {
		return err
	}
	if err := saveIdentity(identityFile, next); err != nil {
		return err
	}
	identityState = next
	return nil
}

func identitySnapshot() IdentityState {
	identityMu.RLock()
	defer identityMu.RUnlock()
	return cloneIdentity(identityState)
}

func saveIdentityLocked() error {
	if identityFile == "" {
		return fmt.Errorf("identity store is not initialized")
	}
	if err := validateIdentity(identityState); err != nil {
		return err
	}
	return saveIdentity(identityFile, identityState)
}

func saveIdentity(path string, state IdentityState) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode identity file: %w", err)
	}
	data = append(data, '\n')
	if len(data) > maxIdentityFileSize {
		return fmt.Errorf("identity file exceeds %d bytes", maxIdentityFileSize)
	}
	file, err := os.CreateTemp(directory, ".identity-*.tmp")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	_, err = io.Copy(file, bytes.NewReader(data))
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(temp, path); err != nil {
		return err
	}
	dir, err := os.Open(directory)
	if err == nil {
		err = dir.Sync()
		_ = dir.Close()
	}
	return err
}

func cloneIdentity(state IdentityState) IdentityState {
	clone := state
	clone.Users = make(map[string]User, len(state.Users))
	for id, user := range state.Users {
		user.GroupIDs = append([]string(nil), user.GroupIDs...)
		clone.Users[id] = user
	}
	clone.Groups = make(map[string]Group, len(state.Groups))
	for id, group := range state.Groups {
		group.Permissions = append([]string(nil), group.Permissions...)
		clone.Groups[id] = group
	}
	clone.Sessions = make(map[string]Session, len(state.Sessions))
	for id, session := range state.Sessions {
		clone.Sessions[id] = session
	}
	clone.Tokens = make(map[string]Token, len(state.Tokens))
	for id, token := range state.Tokens {
		token.Scopes = append([]string(nil), token.Scopes...)
		if token.ExpiresAt != nil {
			value := *token.ExpiresAt
			token.ExpiresAt = &value
		}
		if token.LastUsedAt != nil {
			value := *token.LastUsedAt
			token.LastUsedAt = &value
		}
		if token.RevokedAt != nil {
			value := *token.RevokedAt
			token.RevokedAt = &value
		}
		clone.Tokens[id] = token
	}
	clone.Audit = append([]AuditEvent(nil), state.Audit...)
	return clone
}

func validateIdentity(state IdentityState) error {
	if state.SchemaVersion != IdentitySchemaVersion {
		return fmt.Errorf("unsupported identity schema %d", state.SchemaVersion)
	}
	if state.Users == nil || state.Groups == nil || state.Sessions == nil || state.Tokens == nil {
		return fmt.Errorf("identity state is missing required collections")
	}
	names := make(map[string]string)
	for id, user := range state.Users {
		if id == "" || user.ID != id || user.PasswordHash == "" {
			return fmt.Errorf("identity contains an invalid user")
		}
		if err := ValidateUsername(user.Username); err != nil {
			return fmt.Errorf("user %s: %w", id, err)
		}
		normalized := NormalizeUsername(user.Username)
		if user.Normalized != normalized {
			return fmt.Errorf("user %s has an invalid normalized username", id)
		}
		if existing := names[normalized]; existing != "" && existing != id {
			return fmt.Errorf("duplicate username %s", user.Username)
		}
		names[normalized] = id
		for _, groupID := range user.GroupIDs {
			if _, ok := state.Groups[groupID]; !ok {
				return fmt.Errorf("user %s references missing group %s", id, groupID)
			}
		}
	}
	for id, group := range state.Groups {
		if id == "" || group.ID != id || strings.TrimSpace(group.Name) == "" {
			return fmt.Errorf("identity contains an invalid group")
		}
		for _, permission := range group.Permissions {
			if !IsPermission(permission) {
				return fmt.Errorf("group %s contains unknown permission %s", id, permission)
			}
		}
	}
	for id, session := range state.Sessions {
		if id == "" || session.ID != id || session.SecretHash == "" || session.CSRFHash == "" {
			return fmt.Errorf("identity contains an invalid session")
		}
		if _, ok := state.Users[session.UserID]; !ok {
			return fmt.Errorf("session %s references missing user", id)
		}
	}
	for id, token := range state.Tokens {
		if id == "" || token.ID != id || token.SecretHash == "" || strings.TrimSpace(token.Name) == "" {
			return fmt.Errorf("identity contains an invalid token")
		}
		if _, ok := state.Users[token.OwnerID]; !ok {
			return fmt.Errorf("token %s references missing user", id)
		}
		for _, scope := range token.Scopes {
			if !IsPermission(scope) {
				return fmt.Errorf("token %s contains unknown scope %s", id, scope)
			}
		}
	}
	if state.SetupRequired {
		if len(state.Users) != 0 || state.SetupSecretHash == "" || state.SetupExpiresAt.IsZero() {
			return fmt.Errorf("identity setup state is invalid")
		}
	} else {
		if enabledOwnerCount(state) == 0 {
			return fmt.Errorf("identity does not contain an enabled owner")
		}
		if state.SetupSecretHash != "" || !state.SetupExpiresAt.IsZero() {
			return fmt.Errorf("completed identity setup still contains setup credentials")
		}
	}
	return nil
}

func appendAudit(state *IdentityState, actorID, actorName, action, targetType, targetID string, now time.Time) {
	state.Audit = append(state.Audit, AuditEvent{
		ID:         uuid.NewString(),
		ActorID:    actorID,
		ActorName:  actorName,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  now,
	})
	if overflow := len(state.Audit) - maxAuditEvents; overflow > 0 {
		state.Audit = append([]AuditEvent(nil), state.Audit[overflow:]...)
	}
}
