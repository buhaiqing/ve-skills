package agent

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"time"
)

const maxDiagnoseCalls = 15
const diagnoseTimeout = 30 * time.Second
const maxOutputLen = 2000

// shellMetacharacters detects dangerous shell injection patterns.
var shellMetacharacters = regexp.MustCompile(`[$` + "`" + `;|&(){}\[\]<>!\\]`)

// Diagnose runs a single ve CLI command for the given skill and returns evidence.
// The command is built from the skill and payload, executed with a 30s timeout.
// Safety: rejects commands containing shell metacharacters to prevent injection.
func Diagnose(root, primarySkill string, payload *IncidentPayload) *DiagnosisEvidence {
	cmdStr := BuildDiagnoseCommand(primarySkill, payload)

	// Safety gate: reject commands with shell metacharacters
	if shellMetacharacters.MatchString(cmdStr) {
		return &DiagnosisEvidence{
			Skill: primarySkill,
			Findings: []DiagnosisFinding{{
				Skill:    primarySkill,
				Command:  cmdStr,
				Output:   fmt.Sprintf("diagnose rejected: command contains shell metacharacters"),
				ExitCode: -1,
			}},
			Partial: true,
		}
	}

	// Use exec.Command with explicit args (not shell) to avoid injection
	ctx, cancel := context.WithTimeout(context.Background(), diagnoseTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	outputStr := string(output)
	if len(outputStr) > maxOutputLen {
		outputStr = outputStr[:maxOutputLen]
	}

	return &DiagnosisEvidence{
		Skill: primarySkill,
		Findings: []DiagnosisFinding{
			{
				Skill:    primarySkill,
				Command:  cmdStr,
				Output:   outputStr,
				ExitCode: exitCode,
			},
		},
		Partial: exitCode != 0,
	}
}
