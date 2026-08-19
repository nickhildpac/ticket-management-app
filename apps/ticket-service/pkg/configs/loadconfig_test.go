package configs

import "testing"

func TestIsWeakSecret(t *testing.T) {
	// These placeholders are committed to infra/keycloak/realm-export.json, so
	// they must never pass validation outside local/test.
	weak := []string{
		"secret", "changeme", "change-me", "local-dev-only-secret",
		"ticket-service-dev-secret", "ai-service-dev-secret",
		"  ticket-service-dev-secret  ",
	}
	for _, s := range weak {
		if !isWeakSecret(s) {
			t.Errorf("expected %q to be rejected as weak", s)
		}
	}

	// Empty is not "weak": the admin client is optional, and an unset secret
	// disables role writes rather than misconfiguring them.
	strong := []string{"", "a-real-generated-secret", "Zx9!kd02mfLQ"}
	for _, s := range strong {
		if isWeakSecret(s) {
			t.Errorf("expected %q to be accepted", s)
		}
	}
}

func TestIsLocalOrTest(t *testing.T) {
	for _, env := range []string{"", "local", "test", "LOCAL", " Test "} {
		if !isLocalOrTest(env) {
			t.Errorf("expected %q to count as local/test", env)
		}
	}
	for _, env := range []string{"production", "staging", "development"} {
		if isLocalOrTest(env) {
			t.Errorf("expected %q not to count as local/test", env)
		}
	}
}
