// Package secret implements credential masking for GCL traces.
//
// Faithful Go port of gcl_runner.py's SECRET_PATTERNS / mask_secrets /
// has_credential_leak / detect_credential_fields. Traces MUST never contain
// plaintext credentials — every generator output and command is masked before
// being written to a trace file.
package secret

import "regexp"

// secretPatterns mirrors gcl_runner.SECRET_PATTERNS. They detect plaintext
// Volcengine credentials in trace text. Note: Go's regexp uses RE2 syntax,
// which lacks lookahead, so the Python `(?![<\s])` guard is dropped — the
// `[A-Za-z0-9]{20,}` quantifier already excludes the already-masked `<masked>`
// form (no alphanumerics follow AKLT there).
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`SecretKey\s*=\s*[^<\s][^\s"'']+`),
	regexp.MustCompile(`VOLCENGINE_SECRET_KEY\s*=\s*[^<\s][^\s"'']+`),
	regexp.MustCompile(`AKLT[A-Za-z0-9]{20,}`),
}

// maskSecretPairs mirrors gcl_runner.mask_secrets: replaces the credential
// value (after the '=') with <masked>.
var (
	maskSecretKey = regexp.MustCompile(`(SecretKey\s*=\s*)([^\s"'']+)`)
	maskVolcKey   = regexp.MustCompile(`(VOLCENGINE_SECRET_KEY\s*=\s*)([^\s"'']+)`)
	maskAKLT      = regexp.MustCompile(`(AKLT)([A-Za-z0-9]{20,})`)
)

// MaskSecrets returns text with all detected credential values redacted.
func MaskSecrets(text string) string {
	out := maskSecretKey.ReplaceAllString(text, "${1}<masked>")
	out = maskVolcKey.ReplaceAllString(out, "${1}<masked>")
	out = maskAKLT.ReplaceAllString(out, "${1}<masked>")
	return out
}

// HasCredentialLeak reports whether text contains an unmasked credential.
func HasCredentialLeak(text string) bool {
	for _, p := range secretPatterns {
		if p.MatchString(text) {
			return true
		}
	}
	return false
}

// DetectCredentialFields returns the names of credential fields present in the
// (raw, pre-masking) text. Mirrors gcl_runner.detect_credential_fields.
func DetectCredentialFields(text string) []string {
	var fields []string
	if regexp.MustCompile(`SecretKey\s*=`).MatchString(text) {
		fields = append(fields, "SecretKey")
	}
	if regexp.MustCompile(`VOLCENGINE_SECRET_KEY\s*=`).MatchString(text) {
		fields = append(fields, "VOLCENGINE_SECRET_KEY")
	}
	if regexp.MustCompile(`AKLT[A-Za-z0-9]{20,}`).MatchString(text) {
		fields = append(fields, "AKLT_token")
	}
	return fields
}
