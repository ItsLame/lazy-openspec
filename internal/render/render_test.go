package render

import (
	"strings"
	"testing"

	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tasks"
)

func TestProgressBarWidth(t *testing.T) {
	s := NewSemantic()
	bar := s.ProgressBar(1, 2, 10)
	// Strip ANSI to count glyphs.
	filled := strings.Count(bar, "█")
	empty := strings.Count(bar, "░")
	if filled+empty != 10 {
		t.Fatalf("bar cells = %d, want 10 (%q)", filled+empty, bar)
	}
	if filled != 5 {
		t.Errorf("filled = %d, want 5", filled)
	}
}

func TestScenarioKeywords(t *testing.T) {
	s := NewSemantic()
	raw := "- **WHEN** a user signs in\n- **THEN** a token is returned"
	out := s.Scenario(raw, 60)
	if !strings.Contains(out, "WHEN") || !strings.Contains(out, "THEN") {
		t.Fatalf("missing keywords: %q", out)
	}
	if !strings.Contains(out, "token is returned") {
		t.Errorf("missing body text: %q", out)
	}
}

func TestTasksRendering(t *testing.T) {
	s := NewSemantic()
	groups := tasks.Parse("## 1. Setup\n\n- [ ] 1.1 do thing\n- [x] 1.2 done thing\n")
	out := s.Tasks(groups, 60)
	if !strings.Contains(out, "☐") || !strings.Contains(out, "✔") {
		t.Fatalf("missing checkbox glyphs: %q", out)
	}
}

func TestChangeSpecsBadge(t *testing.T) {
	s := NewSemantic()
	deltas := []openspec.Delta{{
		Spec:      "auth",
		Operation: "ADDED",
		Requirements: []openspec.Requirement{{
			Text:      "The system SHALL do X.",
			Scenarios: []openspec.Scenario{{RawText: "- **WHEN** y\n- **THEN** z"}},
		}},
	}}
	out := s.ChangeSpecs(deltas, 70)
	if !strings.Contains(out, "ADDED") || !strings.Contains(out, "auth") {
		t.Fatalf("missing badge/spec: %q", out)
	}
}

func TestMarkdownRender(t *testing.T) {
	md := NewMarkdown(80)
	out := md.Render("# Title\n\nSome **bold** text.")
	if !strings.Contains(out, "Title") || !strings.Contains(out, "bold") {
		t.Fatalf("markdown render missing content: %q", out)
	}
}
