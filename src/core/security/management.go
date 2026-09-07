package security

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserView struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Enabled   bool      `json:"enabled"`
	GroupIDs  []string  `json:"groupIds"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserUpdate struct {
	Username *string
	Password *string
	Enabled  *bool
	GroupIDs *[]string
}

func ChangeOwnPassword(actor Principal, currentPassword, newPassword string, now time.Time) error {
	hash, err := HashIdentityPassword(newPassword)
	if err != nil {
		return err
	}
	return mutateIdentity(func(state *IdentityState) error {
		user, ok := state.Users[actor.UserID]
		if !ok || !user.Enabled {
			return errors.New("user not found")
		}
		if !verifyIdentityPassword(user.PasswordHash, currentPassword) {
			return errors.New("current password is incorrect")
		}
		user.PasswordHash = hash
		user.UpdatedAt = now
		state.Users[user.ID] = user
		revokeUserCredentials(state, user.ID)
		appendAudit(state, actor.UserID, actor.Username, "password.change", "user", user.ID, now)
		return nil
	})
}

func GetUser(id string) (User, bool) {
	user, ok := identitySnapshot().Users[id]
	return user, ok
}

func ListUsers() []UserView {
	state := identitySnapshot()
	users := make([]UserView, 0, len(state.Users))
	for _, user := range state.Users {
		users = append(users, userView(user))
	}
	sort.Slice(users, func(i, j int) bool {
		return strings.ToLower(users[i].Username) < strings.ToLower(users[j].Username)
	})
	return users
}

func CreateUser(actor Principal, username, password string, groupIDs []string, now time.Time) (UserView, error) {
	if err := ValidateUsername(username); err != nil {
		return UserView{}, err
	}
	hash, err := HashIdentityPassword(password)
	if err != nil {
		return UserView{}, err
	}
	user := User{
		ID: uuid.NewString(), Username: strings.TrimSpace(username), Normalized: NormalizeUsername(username),
		PasswordHash: hash, Enabled: true, CreatedAt: now, UpdatedAt: now,
	}
	err = mutateIdentity(func(state *IdentityState) error {
		if usernameExists(*state, user.Normalized, "") {
			return errors.New("username already exists")
		}
		groups, err := validateGroupGrant(*state, actor, groupIDs)
		if err != nil {
			return err
		}
		user.GroupIDs = groups
		state.Users[user.ID] = user
		appendAudit(state, actor.UserID, actor.Username, "user.create", "user", user.ID, now)
		return nil
	})
	return userView(user), err
}

func UpdateUser(actor Principal, id string, update UserUpdate, now time.Time) (UserView, error) {
	var updated User
	err := mutateIdentity(func(state *IdentityState) error {
		current, ok := state.Users[id]
		if !ok {
			return errors.New("user not found")
		}
		updated = current
		if update.Username != nil {
			if err := ValidateUsername(*update.Username); err != nil {
				return err
			}
			updated.Username = strings.TrimSpace(*update.Username)
			updated.Normalized = NormalizeUsername(*update.Username)
			if usernameExists(*state, updated.Normalized, id) {
				return errors.New("username already exists")
			}
		}
		if update.Enabled != nil {
			updated.Enabled = *update.Enabled
		}
		if update.GroupIDs != nil {
			groups, err := validateGroupGrant(*state, actor, *update.GroupIDs)
			if err != nil {
				return err
			}
			if contains(current.GroupIDs, OwnerGroupID) != contains(groups, OwnerGroupID) && !actor.Permissions[PermissionSecurityManage] {
				return errors.New("changing ownership requires security.manage")
			}
			updated.GroupIDs = groups
		}
		if isEnabledOwner(current) && !isEnabledOwner(updated) && enabledOwnerCount(*state) == 1 {
			return errors.New("the last enabled owner cannot be removed or disabled")
		}
		credentialsChanged := !updated.Enabled
		if update.Password != nil {
			hash, err := HashIdentityPassword(*update.Password)
			if err != nil {
				return err
			}
			updated.PasswordHash = hash
			credentialsChanged = true
		}
		updated.UpdatedAt = now
		state.Users[id] = updated
		if credentialsChanged {
			revokeUserCredentials(state, id)
		}
		appendAudit(state, actor.UserID, actor.Username, "user.update", "user", id, now)
		return nil
	})
	return userView(updated), err
}

func DeleteUser(actor Principal, id string, now time.Time) error {
	return mutateIdentity(func(state *IdentityState) error {
		user, ok := state.Users[id]
		if !ok {
			return errors.New("user not found")
		}
		if isEnabledOwner(user) && enabledOwnerCount(*state) == 1 {
			return errors.New("the last enabled owner cannot be removed")
		}
		revokeUserCredentials(state, id)
		delete(state.Users, id)
		appendAudit(state, actor.UserID, actor.Username, "user.delete", "user", id, now)
		return nil
	})
}

func ListGroups() []Group {
	state := identitySnapshot()
	groups := make([]Group, 0, len(state.Groups))
	for _, group := range state.Groups {
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Name) < strings.ToLower(groups[j].Name)
	})
	return groups
}

func CreateGroup(actor Principal, name, description string, permissions []string, now time.Time) (Group, error) {
	group := Group{
		ID: uuid.NewString(), Name: strings.TrimSpace(name), Description: strings.TrimSpace(description),
		CreatedAt: now, UpdatedAt: now,
	}
	if group.Name == "" || len(group.Name) > 80 {
		return Group{}, errors.New("group name must be between 1 and 80 characters")
	}
	err := mutateIdentity(func(state *IdentityState) error {
		if groupNameExists(*state, group.Name, "") {
			return errors.New("group name already exists")
		}
		granted, err := validatePermissionGrant(actor, permissions)
		if err != nil {
			return err
		}
		group.Permissions = granted
		state.Groups[group.ID] = group
		appendAudit(state, actor.UserID, actor.Username, "group.create", "group", group.ID, now)
		return nil
	})
	return group, err
}

func UpdateGroup(actor Principal, id, name, description string, permissions []string, now time.Time) (Group, error) {
	var updated Group
	err := mutateIdentity(func(state *IdentityState) error {
		group, ok := state.Groups[id]
		if !ok {
			return errors.New("group not found")
		}
		if group.System {
			return errors.New("system groups cannot be changed")
		}
		name = strings.TrimSpace(name)
		if name == "" || len(name) > 80 {
			return errors.New("group name must be between 1 and 80 characters")
		}
		if groupNameExists(*state, name, id) {
			return errors.New("group name already exists")
		}
		granted, err := validatePermissionGrant(actor, permissions)
		if err != nil {
			return err
		}
		group.Name = name
		group.Description = strings.TrimSpace(description)
		group.Permissions = granted
		group.UpdatedAt = now
		state.Groups[id] = group
		updated = group
		appendAudit(state, actor.UserID, actor.Username, "group.update", "group", id, now)
		return nil
	})
	return updated, err
}

func DeleteGroup(actor Principal, id string, now time.Time) error {
	return mutateIdentity(func(state *IdentityState) error {
		group, ok := state.Groups[id]
		if !ok {
			return errors.New("group not found")
		}
		if group.System {
			return errors.New("system groups cannot be deleted")
		}
		delete(state.Groups, id)
		for userID, user := range state.Users {
			user.GroupIDs = without(user.GroupIDs, id)
			state.Users[userID] = user
		}
		appendAudit(state, actor.UserID, actor.Username, "group.delete", "group", id, now)
		return nil
	})
}

func ListTokens(actor Principal) []Token {
	state := identitySnapshot()
	tokens := make([]Token, 0)
	for _, token := range state.Tokens {
		if token.OwnerID != actor.UserID && !actor.Permissions[PermissionSecurityManage] {
			continue
		}
		token.SecretHash = ""
		tokens = append(tokens, token)
	}
	sort.Slice(tokens, func(i, j int) bool { return tokens[i].CreatedAt.After(tokens[j].CreatedAt) })
	return tokens
}

func GetToken(id string) (Token, bool) {
	token, ok := identitySnapshot().Tokens[id]
	return token, ok
}

func ListSessions(actor Principal) []Session {
	state := identitySnapshot()
	sessions := make([]Session, 0)
	for _, session := range state.Sessions {
		if session.UserID != actor.UserID && !actor.Permissions[PermissionSecurityManage] {
			continue
		}
		session.SecretHash = ""
		session.CSRFHash = ""
		sessions = append(sessions, session)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].LastUsedAt.After(sessions[j].LastUsedAt) })
	return sessions
}

func GetSession(id string) (Session, bool) {
	session, ok := identitySnapshot().Sessions[id]
	return session, ok
}

func ListAudit() []AuditEvent {
	state := identitySnapshot()
	events := append([]AuditEvent(nil), state.Audit...)
	sort.Slice(events, func(i, j int) bool { return events[i].CreatedAt.After(events[j].CreatedAt) })
	return events
}

func validateGroupGrant(state IdentityState, actor Principal, requested []string) ([]string, error) {
	groups := make([]string, 0, len(requested))
	seen := make(map[string]bool)
	for _, id := range requested {
		group, ok := state.Groups[id]
		if !ok {
			return nil, errors.New("group does not exist: " + id)
		}
		if id == OwnerGroupID && !actor.Permissions[PermissionSecurityManage] {
			return nil, errors.New("assigning ownership requires security.manage")
		}
		for _, permission := range group.Permissions {
			if !actor.Permissions[permission] {
				return nil, errors.New("group exceeds your access: " + group.Name)
			}
		}
		if !seen[id] {
			seen[id] = true
			groups = append(groups, id)
		}
	}
	return groups, nil
}

func validatePermissionGrant(actor Principal, requested []string) ([]string, error) {
	permissions := make([]string, 0, len(requested))
	seen := make(map[string]bool)
	for _, permission := range requested {
		if !IsPermission(permission) {
			return nil, errors.New("unknown permission: " + permission)
		}
		if !actor.Permissions[permission] {
			return nil, errors.New("permission exceeds your access: " + permission)
		}
		if !seen[permission] {
			seen[permission] = true
			permissions = append(permissions, permission)
		}
	}
	return permissions, nil
}

func usernameExists(state IdentityState, normalized, exceptID string) bool {
	for id, user := range state.Users {
		if id != exceptID && user.Normalized == normalized {
			return true
		}
	}
	return false
}

func groupNameExists(state IdentityState, name, exceptID string) bool {
	for id, group := range state.Groups {
		if id != exceptID && strings.EqualFold(group.Name, name) {
			return true
		}
	}
	return false
}

func enabledOwnerCount(state IdentityState) int {
	count := 0
	for _, user := range state.Users {
		if isEnabledOwner(user) {
			count++
		}
	}
	return count
}

func isEnabledOwner(user User) bool {
	return user.Enabled && contains(user.GroupIDs, OwnerGroupID)
}

func userView(user User) UserView {
	return UserView{
		ID: user.ID, Username: user.Username, Enabled: user.Enabled,
		GroupIDs: append([]string(nil), user.GroupIDs...), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}

func without(values []string, unwanted string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != unwanted {
			result = append(result, value)
		}
	}
	return result
}
