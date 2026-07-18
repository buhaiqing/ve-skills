package agent

type Step int

const (
	StepIngest Step = iota
	StepTriage
	StepDiagnose
	StepPropose
	StepConfirm
	StepExecute
	StepReflexion
	StepDone
)

func (s Step) String() string {
	switch s {
	case StepIngest:
		return "INGEST"
	case StepTriage:
		return "TRIAGE"
	case StepDiagnose:
		return "DIAGNOSE"
	case StepPropose:
		return "PROPOSE"
	case StepConfirm:
		return "CONFIRM"
	case StepExecute:
		return "EXECUTE"
	case StepReflexion:
		return "REFLEXION"
	case StepDone:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

type IncidentPayload struct {
	ProductHint string   `json:"product_hint"`
	Symptom     string   `json:"symptom"`
	TicketID    string   `json:"ticket_id,omitempty"`
	Severity    string   `json:"severity,omitempty"`
	Region      string   `json:"region,omitempty"`
	ResourceIDs []string `json:"resource_ids,omitempty"`
	RawInput    string   `json:"raw_input"`
	Source      string   `json:"source"`
}

type TriageResult struct {
	PrimarySkill    string   `json:"primary_skill"`
	SecondarySkills []string `json:"secondary_skills,omitempty"`
	Confidence      string   `json:"confidence"`
}

type DiagnosisFinding struct {
	Skill    string `json:"skill"`
	Command  string `json:"command"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

type DiagnosisEvidence struct {
	Skill    string             `json:"skill"`
	Findings []DiagnosisFinding `json:"findings"`
	Partial  bool               `json:"partial"`
}
