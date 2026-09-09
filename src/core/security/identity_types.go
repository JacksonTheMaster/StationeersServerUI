package security

import "time"

const IdentitySchemaVersion = 1
const OwnerGroupID = "system-owner"

type IdentityState struct {
	SchemaVersion int                `json:"schemaVersion"`
	SetupRequired bool               `json:"setupRequired"`
	Users         map[string]User    `json:"users"`
	Groups        map[string]Group   `json:"groups"`
	Sessions      map[string]Session `json:"sessions"`
	Tokens        map[string]Token   `json:"tokens"`
	Audit         []AuditEvent       `json:"audit"`
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Normalized   string    `json:"normalized"`
	PasswordHash string    `json:"passwordHash"`
	Enabled      bool      `json:"enabled"`
	GroupIDs     []string  `json:"groupIds"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	System      bool      `json:"system"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Session struct {
	ID                string    `json:"id"`
	SecretHash        string    `json:"secretHash"`
	CSRFHash          string    `json:"csrfHash"`
	UserID            string    `json:"userId"`
	CreatedAt         time.Time `json:"createdAt"`
	LastUsedAt        time.Time `json:"lastUsedAt"`
	IdleExpiresAt     time.Time `json:"idleExpiresAt"`
	AbsoluteExpiresAt time.Time `json:"absoluteExpiresAt"`
}

type Token struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	SecretHash string     `json:"secretHash"`
	OwnerID    string     `json:"ownerId"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

type AuditEvent struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actorId,omitempty"`
	ActorName  string    `json:"actorName"`
	Action     string    `json:"action"`
	TargetType string    `json:"targetType"`
	TargetID   string    `json:"targetId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Principal struct {
	UserID       string
	Username     string
	CredentialID string
	Credential   string
	Permissions  map[string]bool
}
