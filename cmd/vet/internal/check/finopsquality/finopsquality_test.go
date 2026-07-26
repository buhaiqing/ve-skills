package finopsquality

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// RED: A quality finops.md with ≥20 lines, pricing, and ≥2 optimizations should pass.
func TestCheckSkillQualityPass(t *testing.T) {
	dir := t.TempDir()
	content := `# FinOps — ECS

## Pricing
ECS instances are billed per hour. See ve ecs DescribePrice.

## Cost Optimization

### Right-Sizing
Recommend downgrading instances with CPU avg < 15%.

### Idle Detection
Recommend releasing unattached EBS volumes.

### Spot Instances
Use spot for stateless workloads.

## Billing
Query billing via ve billing DescribeBillDetail.
`
	// 10 lines — required needs ≥20. Make it longer.
	for i := 0; i < 12; i++ {
		content += "\n## Section " + itoa(i) + "\nSome content here."
	}
	writeFile(t, filepath.Join(dir, "ve-ecs-ops", "references", "advanced", "finops.md"), content)

	r := checkSkill(dir, "ve-ecs-ops")
	if !r.HasFinOps {
		t.Fatal("expected HasFinOps true")
	}
	if r.LineCount < 20 {
		t.Errorf("expected ≥20 lines for required tier, got %d", r.LineCount)
	}
	if r.HasPricing == "" {
		t.Error("expected HasPricing to detect pricing keywords")
	}
	if !r.HasOptimizations {
		t.Error("expected HasOptimizations true for ≥2 optimizations")
	}
	if r.OK {
		t.Log("PASS: quality finops.md passed all checks")
	}
}

// RED: A template/empty finops.md should fail quality checks.
func TestCheckSkillQualityFail(t *testing.T) {
	dir := t.TempDir()
	// Minimal template — no pricing, no optimizations
	content := `# FinOps — ECS

Template content. Replace with product-specific details.
`
	writeFile(t, filepath.Join(dir, "ve-ecs-ops", "references", "advanced", "finops.md"), content)

	r := checkSkill(dir, "ve-ecs-ops")
	if !r.HasFinOps {
		t.Fatal("expected HasFinOps true (file exists)")
	}
	if r.LineCount >= 20 {
		t.Errorf("expected <20 lines for template finops.md, got %d", r.LineCount)
	}
	if r.HasPricing != "" {
		t.Errorf("expected no pricing detection for template, got %q", r.HasPricing)
	}
	if r.HasOptimizations {
		t.Error("expected HasOptimizations false for template")
	}
	if r.OK {
		t.Error("expected OK false for template finops.md")
	}
}

// RED: Missing finops.md should report HasFinOps=false.
func TestCheckSkillMissing(t *testing.T) {
	dir := t.TempDir()
	r := checkSkill(dir, "ve-ecs-ops")
	if r.HasFinOps {
		t.Error("expected HasFinOps false when finops.md missing")
	}
}

// RED: CheckDir should produce correct aggregate report.
func TestCheckDirReport(t *testing.T) {
	dir := t.TempDir()
	// Create quality finops.md for one R+R skill
	content := `# FinOps — ECS

## Pricing
ECS pricing per instance type.

## Cost Optimization

### Right-Sizing
Right-size idle instances.

### Spot Usage
Use spot instances.

## Billing
Billing via DescribeBill.
`
	for i := 0; i < 14; i++ {
		content += "\n## S" + itoa(i) + "\nMore content."
	}
	writeFile(t, filepath.Join(dir, "ve-ecs-ops", "references", "advanced", "finops.md"), content)
	// Template finops.md for another skill
	writeFile(t, filepath.Join(dir, "ve-redis-ops", "references", "advanced", "finops.md"), "# FinOps — Redis\nTemplate.\n")

	rep := CheckDir(dir)
	if rep.TotalSkills <= 0 {
		t.Fatal("expected >0 total skills")
	}
	// Quality results
	found := false
	for _, q := range rep.Quality {
		if q.Skill == "ve-ecs-ops" && q.OK {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ve-ecs-ops to pass quality checks")
	}
}

func itoa(n int) string {
	return string(rune('0' + n))
}