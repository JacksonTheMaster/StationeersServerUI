package config

import "testing"

func TestGetOptionalIntPreservesExplicitZero(t *testing.T) {
	zero := 0
	if got := getOptionalInt(&zero, "TEST_OPTIONAL_INT", 42); got != 0 {
		t.Fatalf("getOptionalInt() = %d, want explicit zero", got)
	}
}

func TestGetOptionalIntFallbackOrder(t *testing.T) {
	t.Setenv("TEST_OPTIONAL_INT", "17")
	if got := getOptionalInt(nil, "TEST_OPTIONAL_INT", 42); got != 17 {
		t.Fatalf("getOptionalInt() = %d, want environment value 17", got)
	}

	t.Setenv("TEST_OPTIONAL_INT", "")
	if got := getOptionalInt(nil, "TEST_OPTIONAL_INT", 42); got != 42 {
		t.Fatalf("getOptionalInt() = %d, want default value 42", got)
	}
}
