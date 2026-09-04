package configchanger

import "testing"

func TestDiscordAdminRoleValidation(t *testing.T) {
	for _, role := range []string{"", "123456789012345678", "18446744073709551615"} {
		if err := validateDiscordAdminRole(role); err != nil {
			t.Fatalf("valid role %q: %v", role, err)
		}
	}
	for _, role := range []string{"0", "<@&123>", "Admins", " 123", "+123", "-1", "1.5", "18446744073709551616"} {
		if err := validateDiscordAdminRole(role); err == nil {
			t.Fatalf("accepted invalid role %q", role)
		}
	}
}
