package security

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestInitializeIdentityMigratesUsersAndDropsAPIKeys(t *testing.T) {
	useIdentityTestDirectory(t)
	now := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	legacyHash, err := bcrypt.GenerateFromPassword([]byte("old-but-valid-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	secret, err := InitializeIdentity(map[string]string{
		"Jackson":                          string(legacyHash),
		"apikey-this-must-not-be-imported": string(legacyHash),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if secret != "" || SetupRequired() {
		t.Fatal("legacy user migration should complete setup")
	}
	users := ListUsers()
	if len(users) != 1 || users[0].Username != "Jackson" {
		t.Fatalf("unexpected migrated users: %#v", users)
	}
	if _, err := AuthenticateUser("jackson", "old-but-valid-password"); err != nil {
		t.Fatalf("migrated user could not log in: %v", err)
	}
	state := identitySnapshot()
	if !strings.HasPrefix(state.Users[users[0].ID].PasswordHash, "$argon2id$") {
		t.Fatal("legacy bcrypt password was not upgraded after login")
	}
}

func TestOwnerBootstrapSessionAndToken(t *testing.T) {
	useIdentityTestDirectory(t)
	now := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	secret, err := InitializeIdentity(nil, now)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := BootstrapOwner(secret, "admin", "correct horse battery staple", now)
	if err != nil {
		t.Fatal(err)
	}

	session, credential, err := CreateSession(owner.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	principal, authenticated, err := AuthenticateSession(credential.Value, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if authenticated.ID != session.ID || !principal.Permissions[PermissionSecurityManage] {
		t.Fatal("session did not resolve the owner permissions")
	}
	if !ValidateSessionCSRF(authenticated, credential.CSRF) || ValidateSessionCSRF(authenticated, "wrong") {
		t.Fatal("session CSRF validation returned the wrong result")
	}

	token, value, err := CreateToken(owner.ID, "automation", []string{PermissionServerView}, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	tokenPrincipal, authenticatedToken, err := AuthenticateToken(value, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if authenticatedToken.ID != token.ID || !tokenPrincipal.Permissions[PermissionServerView] || tokenPrincipal.Permissions[PermissionServerControl] {
		t.Fatal("token scopes were not enforced")
	}
}

func TestChangeOwnPasswordRevokesExistingCredentials(t *testing.T) {
	useIdentityTestDirectory(t)
	now := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	secret, err := InitializeIdentity(nil, now)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := BootstrapOwner(secret, "admin", "correct horse battery staple", now)
	if err != nil {
		t.Fatal(err)
	}
	_, credential, err := CreateSession(owner.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	principal, _, err := AuthenticateSession(credential.Value, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := ChangeOwnPassword(principal, "correct horse battery staple", "a much better password", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AuthenticateSession(credential.Value, now.Add(2*time.Minute)); err == nil {
		t.Fatal("old session survived the password change")
	}
	if _, err := AuthenticateUser("admin", "a much better password"); err != nil {
		t.Fatal("new password did not work")
	}
}

func TestFailedIdentityMutationDoesNotLeakIntoMemory(t *testing.T) {
	useIdentityTestDirectory(t)
	now := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	secret, err := InitializeIdentity(nil, now)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := BootstrapOwner(secret, "admin", "correct horse battery staple", now)
	if err != nil {
		t.Fatal(err)
	}

	err = mutateIdentity(func(state *IdentityState) error {
		user := state.Users[owner.ID]
		user.Enabled = false
		state.Users[owner.ID] = user
		return errors.New("nope")
	})
	if err == nil {
		t.Fatal("expected mutation to fail")
	}
	user, ok := GetUser(owner.ID)
	if !ok || !user.Enabled {
		t.Fatal("failed mutation changed the live identity state")
	}
}

func useIdentityTestDirectory(t *testing.T) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(workingDirectory)
		identityMu.Lock()
		identityState = IdentityState{}
		identityFile = ""
		identityMu.Unlock()
	})
	identityMu.Lock()
	identityState = IdentityState{}
	identityFile = filepath.Join("SSUI", "security", "identity.json")
	identityMu.Unlock()
}
