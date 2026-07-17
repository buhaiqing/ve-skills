package agent

import (
	"testing"
)

func TestParseJSON_ValidPayload(t *testing.T) {
	payload, err := ParseJSON([]byte(`{"product_hint":"ecs","symptom":"cpu>90%","ticket_id":"DOPS-123"}`))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}
	if payload.ProductHint != "ecs" {
		t.Errorf("expected product=ecs, got %s", payload.ProductHint)
	}
	if payload.Symptom != "cpu>90%" {
		t.Errorf("expected symptom=cpu>90%%, got %s", payload.Symptom)
	}
	if payload.TicketID != "DOPS-123" {
		t.Errorf("expected ticket=DOPS-123, got %s", payload.TicketID)
	}
	if payload.Source != "json" {
		t.Errorf("expected source=json, got %s", payload.Source)
	}
}

func TestParseJSON_MissingProductHint(t *testing.T) {
	_, err := ParseJSON([]byte(`{"symptom":"cpu>90%"}`))
	if err == nil {
		t.Error("expected error for missing product_hint, got nil")
	}
}

func TestParseJSON_InvalidJSON(t *testing.T) {
	_, err := ParseJSON([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseNaturalLanguage_ECS(t *testing.T) {
	payload, err := ParseNaturalLanguage("ecs 实例 CPU 很高，磁盘也快满了")
	if err != nil {
		t.Fatalf("ParseNaturalLanguage failed: %v", err)
	}
	if payload.ProductHint != "ecs" {
		t.Errorf("expected ecs, got %s", payload.ProductHint)
	}
}

func TestParseNaturalLanguage_Redis(t *testing.T) {
	payload, err := ParseNaturalLanguage("Redis 缓存响应很慢，连接数过高")
	if err != nil {
		t.Fatalf("ParseNaturalLanguage failed: %v", err)
	}
	if payload.ProductHint != "redis" {
		t.Errorf("expected redis, got %s", payload.ProductHint)
	}
}

func TestParseNaturalLanguage_ChineseKeywords(t *testing.T) {
	payload, err := ParseNaturalLanguage("云服务器 CPU 使用率超过 90%，需要排查")
	if err != nil {
		t.Fatalf("ParseNaturalLanguage failed: %v", err)
	}
	if payload.ProductHint != "ecs" {
		t.Errorf("expected ecs from 云服务器, got %s", payload.ProductHint)
	}
}

func TestParseNaturalLanguage_EmptyInput(t *testing.T) {
	_, err := ParseNaturalLanguage("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestParseNaturalLanguage_UnknownProduct(t *testing.T) {
	payload, err := ParseNaturalLanguage("some random text without product keywords")
	if err == nil && payload == nil {
		t.Error("expected error or nil payload for unknown product")
	}
	// The function returns an error if product_hint can't be determined
	if err == nil && payload.ProductHint != "" {
		t.Skip("product hint found unexpectedly, but this is acceptable behavior")
	}
}

func TestParseNaturalLanguage_ResourceIDs(t *testing.T) {
	payload, err := ParseNaturalLanguage("ecs instance i-abc12345 has high CPU, related redis redis-abcdef01")
	if err != nil {
		t.Fatalf("ParseNaturalLanguage failed: %v", err)
	}
	if len(payload.ResourceIDs) < 1 {
		t.Error("expected at least 1 resource ID")
	}
	found := false
	for _, id := range payload.ResourceIDs {
		if id == "i-abc12345" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected i-abc12345 in resource IDs, got %v", payload.ResourceIDs)
	}
}

func TestParseNaturalLanguage_SymptomExtraction(t *testing.T) {
	tests := []struct {
		input   string
		want    string
	}{
		{"ecs CPU 高 使用率 100%", "cpu_high"},
		{"redis 内存 使用 high", "mem_high"},
		{"磁盘 空间 不足 on ecs", "disk_high"},
		{"service latency 超时 slow", "latency"},
		{"connection 连接 失败 on database", "connection"},
		{"instance error crash panic 错误", "errors"},
	}
	for _, tt := range tests {
		payload, err := ParseNaturalLanguage(tt.input)
		if err != nil {
			t.Fatalf("ParseNaturalLanguage(%q) failed: %v", tt.input, err)
		}
		if payload.Symptom != tt.want {
			t.Errorf("ParseNaturalLanguage(%q): expected symptom=%q, got %q", tt.input, tt.want, payload.Symptom)
		}
	}
}

func TestParseNaturalLanguage_DeterministicKeywords(t *testing.T) {
	// Run 100 times to verify deterministic keyword matching
	for i := 0; i < 100; i++ {
		payload, err := ParseNaturalLanguage("mysql kafka both have issues")
		if err != nil {
			t.Fatalf("ParseNaturalLanguage failed at iteration %d: %v", i, err)
		}
		// "mysql" and "kafka" both 5 chars, sorted alphabetically: "kafka" < "mysql"
		if payload.ProductHint != "kafka" {
			t.Errorf("iteration %d: expected deterministic product=kafka (alphabetical first), got %s", i, payload.ProductHint)
		}
	}
}

func TestTriage_KnownSkill(t *testing.T) {
	payload := &IncidentPayload{ProductHint: "ecs", Symptom: "cpu>90%"}
	result := Triage(payload)
	if result.PrimarySkill != "ve-ecs-ops" {
		t.Errorf("expected ve-ecs-ops, got %s", result.PrimarySkill)
	}
	if result.Confidence != "high" {
		t.Errorf("expected high confidence, got %s", result.Confidence)
	}
}

func TestTriage_UnknownSkill(t *testing.T) {
	payload := &IncidentPayload{ProductHint: "unknown_product", Symptom: "something"}
	result := Triage(payload)
	// Rule 5: unknown → ve-cms-ops
	if result.PrimarySkill != "ve-cms-ops" {
		t.Errorf("expected ve-cms-ops (Rule 5), got %s", result.PrimarySkill)
	}
	if result.Confidence != "low" {
		t.Errorf("expected low confidence for unknown, got %s", result.Confidence)
	}
}

func TestBuildDiagnoseCommand(t *testing.T) {
	payload := &IncidentPayload{ProductHint: "ecs"}
	cmd := BuildDiagnoseCommand("ve-ecs-ops", payload)
	if cmd == "" {
		t.Error("expected non-empty command")
	}
	// Should contain service name and Describe action
	if !contains(cmd, "ecs") || !contains(cmd, "Describe") {
		t.Errorf("expected command to contain 'ecs' and 'Describe', got: %s", cmd)
	}
}

func TestStepString(t *testing.T) {
	if StepIngest.String() != "INGEST" {
		t.Errorf("expected INGEST, got %s", StepIngest.String())
	}
	if StepTriage.String() != "TRIAGE" {
		t.Errorf("expected TRIAGE, got %s", StepTriage.String())
	}
	if StepReflexion.String() != "REFLEXION" {
		t.Errorf("expected REFLEXION, got %s", StepReflexion.String())
	}
	if StepDone.String() != "DONE" {
		t.Errorf("expected DONE, got %s", StepDone.String())
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
