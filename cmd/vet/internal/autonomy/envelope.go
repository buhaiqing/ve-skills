package autonomy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Domain defines a single autonomous domain within the envelope.
type Domain struct {
	Name        string   `yaml:"name"`
	Skills      []string `yaml:"skills"`
	Symptoms    []string `yaml:"symptoms"`
	BlastRadius string   `yaml:"blast_radius"`
	SLORef      string   `yaml:"slo_ref"`
}

// Envelope defines the full set of autonomous domains.
type Envelope struct {
	Domains []Domain `yaml:"domains"`
}

// LoadEnvelope reads and parses an envelope YAML file.
func LoadEnvelope(path string) (*Envelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading envelope file: %w", err)
	}

	var env Envelope
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parsing envelope YAML: %w", err)
	}

	return &env, nil
}

// InEnvelope checks if (skill, symptom, blastRadius) is allowed within the envelope.
func (e *Envelope) InEnvelope(skill, symptom, blastRadius string) bool {
	for _, d := range e.Domains {
		if d.matchesSkill(skill) && d.matchesSymptom(symptom) {
			// Check blast radius cap
			if d.BlastRadius == "single" && blastRadius != "single" {
				return false
			}
			return true
		}
	}
	return false
}

// ListSkills returns all unique skills covered by the envelope.
func (e *Envelope) ListSkills() []string {
	seen := make(map[string]bool)
	var skills []string
	for _, d := range e.Domains {
		for _, s := range d.Skills {
			if !seen[s] {
				seen[s] = true
				skills = append(skills, s)
			}
		}
	}
	return skills
}

// GetDomain returns the first matching domain for (skill, symptom).
func (e *Envelope) GetDomain(skill, symptom string) *Domain {
	for i := range e.Domains {
		if e.Domains[i].matchesSkill(skill) && e.Domains[i].matchesSymptom(symptom) {
			return &e.Domains[i]
		}
	}
	return nil
}

// matchesSkill checks if the domain covers the given skill.
func (d *Domain) matchesSkill(skill string) bool {
	for _, s := range d.Skills {
		if s == skill {
			return true
		}
	}
	return false
}

// matchesSymptom checks if the domain covers the given symptom.
func (d *Domain) matchesSymptom(symptom string) bool {
	for _, s := range d.Symptoms {
		if s == symptom {
			return true
		}
	}
	return false
}
