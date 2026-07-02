package configs

import "testing"

func TestIsWeakJWTSecret(t *testing.T) {
	for _, secret := range []string{"", "secret", "changeme", "change-me", "local-dev-only-secret"} {
		if !isWeakJWTSecret(secret) {
			t.Fatalf("expected %q to be weak", secret)
		}
	}

	if isWeakJWTSecret("a-production-secret-with-enough-entropy") {
		t.Fatal("expected strong-looking secret to pass weak-value check")
	}
}

func TestIsLocalOrTest(t *testing.T) {
	for _, env := range []string{"", "local", "test"} {
		if !isLocalOrTest(env) {
			t.Fatalf("expected %q to be local/test", env)
		}
	}

	if isLocalOrTest("production") {
		t.Fatal("expected production not to be local/test")
	}
}
