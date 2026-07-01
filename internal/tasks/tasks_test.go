package tasks

import "testing"

const sample = `## 1. Backend auth

- [ ] 1.1 Add user model
- [x] 1.2 Hash passwords

## 2. Frontend

- [ ] 2.1 Build login form
`

func TestParse(t *testing.T) {
	groups := Parse(sample)
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d", len(groups))
	}
	if groups[0].Title != "Backend auth" || groups[0].Number != "1" {
		t.Errorf("group 0 = %+v", groups[0])
	}
	if got := groups[0].Total(); got != 2 {
		t.Errorf("group 0 total = %d, want 2", got)
	}
	if got := groups[0].Completed(); got != 1 {
		t.Errorf("group 0 completed = %d, want 1", got)
	}
	if !groups[0].Tasks[1].Done {
		t.Errorf("task 1.2 should be done")
	}
}

func TestProgress(t *testing.T) {
	done, total := Progress(sample)
	if done != 1 || total != 3 {
		t.Fatalf("progress = %d/%d, want 1/3", done, total)
	}
}

func TestToggleFlipsAndPreservesBytes(t *testing.T) {
	out, nowDone, err := Toggle(sample, "1.1")
	if err != nil {
		t.Fatalf("toggle: %v", err)
	}
	if !nowDone {
		t.Errorf("1.1 should now be done")
	}
	// Only the single marker should change: toggling back yields the original.
	back, nowDone2, err := Toggle(out, "1.1")
	if err != nil {
		t.Fatalf("toggle back: %v", err)
	}
	if nowDone2 {
		t.Errorf("1.1 should be undone")
	}
	if back != sample {
		t.Errorf("round-trip toggle did not restore original bytes:\n%q\n!=\n%q", back, sample)
	}
}

func TestToggleUnknownTaskErrors(t *testing.T) {
	if _, _, err := Toggle(sample, "9.9"); err == nil {
		t.Fatalf("expected error toggling unknown task")
	}
}
