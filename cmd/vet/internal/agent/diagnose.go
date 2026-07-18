package agent

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const maxDiagnoseCalls = 15
const diagnoseTimeout = 30 * time.Second
const maxOutputLen = 2000

// shellMetacharacters detects dangerous shell injection patterns.
var shellMetacharacters = regexp.MustCompile(`[$` + "`" + `;|&(){}\[\]<>!\\]`)

func Diagnose(root, primarySkill string, payload *IncidentPayload) *DiagnosisEvidence {
	args := BuildDiagnoseArgs(primarySkill, payload)

	// Safety gate: reject args containing shell metacharacters
	for _, arg := range args {
		if shellMetacharacters.MatchString(arg) {
			return &DiagnosisEvidence{
				Skill: primarySkill,
				Findings: []DiagnosisFinding{{
					Skill:    primarySkill,
					Command:  "ve " + strings.Join(args, " "),
					Output:   "diagnose rejected: argument contains shell metacharacters",
					ExitCode: -1,
				}},
				Partial: true,
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), diagnoseTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ve", args...)

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
				Command:  "ve " + strings.Join(args, " "),
				Output:   outputStr,
				ExitCode: exitCode,
			},
		},
		Partial: exitCode != 0,
	}
}
