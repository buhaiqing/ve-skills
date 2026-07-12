package secret

import "testing"

func TestMaskSecrets(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"SecretKey=EXAMPLE_VALUE", "SecretKey=<masked>"},
		{"VOLCENGINE_SECRET_KEY=EXAMPLE_VALUE", "VOLCENGINE_SECRET_KEY=<masked>"},
		{"AKLT-PLACEHOLDER-TOKEN", "AKLT<masked>"},
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
	if !HasCredentialLeak("AKLT-PLACEHOLDER-TOKEN") {
		t.Error("expected leak for AKLT token")
	}
	if HasCredentialLeak("SecretKey=<masked>") {
		t.Error("masked value must not be a leak")
	}
	if HasCredentialLeak("plain text") {
		t.Error("plain text must not be a leak")
	}
}

func TestDetectCredentialFields(t *testing.T) {
	f := DetectCredentialFields("SecretKey=abc VOLCENGINE_SECRET_KEY=xyz AKLT-PLACEHOLDER-TOKEN")
	want := map[string]bool{"SecretKey": true, "VOLCENGINE_SECRET_KEY": true, "AKLT_token": true}
	for _, name := range []string{"SecretKey", "VOLCENGINE_SECRET_KEY", "AKLT_token"} {
		if !want[name] {
			t.Errorf("expected field %s detected", name)
		}
		if !contains(f, name) {
			t.Errorf("field %s not in %v", name, f)
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
