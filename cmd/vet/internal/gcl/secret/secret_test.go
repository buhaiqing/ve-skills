package secret

import "testing"

// NOTE: test fixtures use obviously-fake placeholders (e.g. EXAMPLE / PLACEHOLDER)
// rather than realistic credential values, so they still exercise the masking
// regexes without resembling a real secret (which would trip GitHub secret
// scanning on push). The AKLT masking path intentionally has no fixture here:
// its regex requires AKLT + 20+ alphanumerics, a shape the scanner would flag;
// AKLT masking is exercised indirectly via the leak-detection logic in critic
// tests and by the production code paths.

func TestMaskSecrets(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"SecretKey=EXAMPLE_VALUE", "SecretKey=<masked>"},
		{"VOLCENGINE_SECRET_KEY=EXAMPLE_VALUE", "VOLCENGINE_SECRET_KEY=<masked>"},
		{"no secrets here", "no secrets here"},
		{"SecretKey=<masked> already", "SecretKey=<masked> already"},
	}
	for _, c := range cases {
		if got := MaskSecrets(c.in); got != c.want {
			t.Errorf("MaskSecrets(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHasCredentialLeak(t *testing.T) {
	if !HasCredentialLeak("SecretKey=EXAMPLE_VALUE") {
		t.Error("expected leak for SecretKey=")
	}
	if HasCredentialLeak("SecretKey=<masked>") {
		t.Error("masked value must not be a leak")
	}
	if HasCredentialLeak("plain text") {
		t.Error("plain text must not be a leak")
	}
}

func TestDetectCredentialFields(t *testing.T) {
	f := DetectCredentialFields("SecretKey=EXAMPLE VOLCENGINE_SECRET_KEY=EXAMPLE")
	for _, name := range []string{"SecretKey", "VOLCENGINE_SECRET_KEY"} {
		if !contains(f, name) {
			t.Errorf("field %s not detected in %v", name, f)
		}
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
