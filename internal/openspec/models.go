// Package openspec is the data-access layer for lazy-openspec. It shells out to
// the installed `openspec` CLI with its `--json` flag and decodes the structured
// output into typed models, rather than parsing the openspec/ markdown tree.
package openspec

import "encoding/json"

// firstNonEmpty returns the first non-empty string in vals, or "".
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Root is the resolved OpenSpec root reported by most CLI commands.
type Root struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}

// ChangeSummary is one entry from `openspec list --json`.
type ChangeSummary struct {
	Name           string `json:"name"`
	CompletedTasks int    `json:"completedTasks"`
	TotalTasks     int    `json:"totalTasks"`
	LastModified   string `json:"lastModified"`
	Status         string `json:"status"`
}

// Lifecycle groups a change into the buckets lazy-openspec renders under.
type Lifecycle int

const (
	// Draft: no tasks defined yet.
	Draft Lifecycle = iota
	// Active: some but not all tasks complete.
	Active
	// Completed: every task complete.
	Completed
)

// Lifecycle derives the lifecycle bucket from task counts, matching the rules
// the CLI's own dashboard uses (0 tasks => draft, all done => completed).
func (c ChangeSummary) Lifecycle() Lifecycle {
	switch {
	case c.TotalTasks == 0:
		return Draft
	case c.CompletedTasks >= c.TotalTasks:
		return Completed
	default:
		return Active
	}
}

// Percent is the completion percentage (0-100), guarding against divide-by-zero.
func (c ChangeSummary) Percent() int {
	if c.TotalTasks == 0 {
		return 0
	}
	return int(float64(c.CompletedTasks) / float64(c.TotalTasks) * 100)
}

// ChangeList is the payload of `openspec list --json`.
type ChangeList struct {
	Changes []ChangeSummary `json:"changes"`
	Root    Root            `json:"root"`
}

// SpecSummary is one entry from `openspec list --specs --json`. Name is decoded
// tolerantly: the CLI keys specs by `id` (openspec 1.5.0), while older/other
// versions may use `name`; either populates Name so the identity is never blank.
type SpecSummary struct {
	Name             string `json:"id"`
	RequirementCount int    `json:"requirementCount"`
}

// UnmarshalJSON accepts the spec identifier under either `id` or `name`.
func (s *SpecSummary) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		RequirementCount int    `json:"requirementCount"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Name = firstNonEmpty(raw.ID, raw.Name)
	s.RequirementCount = raw.RequirementCount
	return nil
}

// SpecList is the payload of `openspec list --specs --json`.
type SpecList struct {
	Specs []SpecSummary `json:"specs"`
	Root  Root          `json:"root"`
}

// ArtifactStatus is one artifact entry from `openspec status --json`.
type ArtifactStatus struct {
	ID          string   `json:"id"`
	OutputPath  string   `json:"outputPath"`
	Status      string   `json:"status"`
	MissingDeps []string `json:"missingDeps,omitempty"`
}

// Done reports whether the artifact is complete.
func (a ArtifactStatus) Done() bool { return a.Status == "done" }

// Status is the payload of `openspec status --change <name> --json`.
type Status struct {
	ChangeName    string           `json:"changeName"`
	SchemaName    string           `json:"schemaName"`
	ChangeRoot    string           `json:"changeRoot"`
	IsComplete    bool             `json:"isComplete"`
	ApplyRequires []string         `json:"applyRequires"`
	Artifacts     []ArtifactStatus `json:"artifacts"`
}

// Scenario is a single WHEN/THEN scenario captured as raw markdown text.
type Scenario struct {
	RawText string `json:"rawText"`
}

// Requirement is a normative requirement plus its scenarios.
type Requirement struct {
	Text      string     `json:"text"`
	Scenarios []Scenario `json:"scenarios"`
}

// Delta is one delta operation from `openspec show <change> --json`.
type Delta struct {
	Spec         string        `json:"spec"`
	Operation    string        `json:"operation"`
	Description  string        `json:"description"`
	Requirements []Requirement `json:"requirements"`
}

// ChangeDetail is the payload of `openspec show <change> --json`.
type ChangeDetail struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	DeltaCount int     `json:"deltaCount"`
	Deltas     []Delta `json:"deltas"`
}

// SpecDetail is the payload of `openspec spec show <id> --json`. The field names
// are decoded tolerantly; different CLI versions expose id/name/title and
// requirements.
type SpecDetail struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Requirements []Requirement `json:"requirements"`
}

// UnmarshalJSON accepts the spec identifier under `id` or `name`, and a display
// name under `name` or `title` (openspec 1.5.0 emits `id` + `title`).
func (s *SpecDetail) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID           string        `json:"id"`
		Name         string        `json:"name"`
		Title        string        `json:"title"`
		Requirements []Requirement `json:"requirements"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.ID = firstNonEmpty(raw.ID, raw.Name)
	s.Name = firstNonEmpty(raw.Name, raw.Title, raw.ID)
	s.Requirements = raw.Requirements
	return nil
}
