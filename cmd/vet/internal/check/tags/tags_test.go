package tags

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

func TestCheckSkillMissingTags(t *testing.T) {
	dir := t.TempDir()
	content := `# ve-ecs-ops

` + "```bash" + `
ve ecs RunInstances --InstanceType ecs.g1.large --ImageId image-xxx
` + "```" + `
`
	writeFile(t, filepath.Join(dir, "ve-ecs-ops", "SKILL.md"), content)
	r := checkSkill(dir, "ve-ecs-ops")
	if r.HasTags {
		t.Error("expected HasTags=false when Create command lacks --Tags")
	}
	if r.OK {
		t.Error("expected OK=false for missing --Tags")
	}
}

func TestCheckSkillHasTags(t *testing.T) {
	dir := t.TempDir()
	content := `# ve-ecs-ops

` + "```bash" + `
ve ecs RunInstances --InstanceType ecs.g1.large --Tags "[{\"Key\":\"cost-center\",\"Value\":\"engineering\"}]"
` + "```" + `
`
	writeFile(t, filepath.Join(dir, "ve-ecs-ops", "SKILL.md"), content)
	r := checkSkill(dir, "ve-ecs-ops")
	if !r.HasTags {
		t.Error("expected HasTags=true when Create command includes --Tags")
	}
	if !r.OK {
		t.Error("expected OK=true for compliant SKILL.md")
	}
}

func TestCheckSkillNoCreate(t *testing.T) {
	dir := t.TempDir()
	content := `# ve-readonly-ops

` + "```bash" + `
ve ecs DescribeInstances --InstanceId i-xxx
` + "```" + `
`
	writeFile(t, filepath.Join(dir, "ve-readonly-ops", "SKILL.md"), content)
	r := checkSkill(dir, "ve-readonly-ops")
	if !r.OK {
		t.Error("expected OK=true when no Create commands exist")
	}
}

func TestCheckDirReport(t *testing.T) {
	dir := t.TempDir()
	contentOK := `# ve-ecs-ops

` + "```bash" + `
ve ecs RunInstances --InstanceType ecs.g1.large --Tags "[{\"Key\":\"env\",\"Value\":\"prod\"}]"
` + "```" + `
`
	writeFile(t, filepath.Join(dir, "ve-ecs-ops", "SKILL.md"), contentOK)
	contentFail := `# ve-redis-ops

` + "```bash" + `
ve redis CreateDBInstance --InstanceName myredis
` + "```" + `
`
	writeFile(t, filepath.Join(dir, "ve-redis-ops", "SKILL.md"), contentFail)

	rep := CheckDir(dir)
	if rep.TotalSkills <= 0 {
		t.Fatal("expected >0 total skills")
	}
	foundPass := false
	foundFail := false
	for _, q := range rep.Quality {
		if q.Skill == "ve-ecs-ops" && q.OK {
			foundPass = true
		}
		if q.Skill == "ve-redis-ops" && !q.OK && !q.HasTags {
			foundFail = true
		}
	}
	if !foundPass {
		t.Error("expected ve-ecs-ops to pass")
	}
	if !foundFail {
		t.Error("expected ve-redis-ops to fail with missing Tags")
	}
}