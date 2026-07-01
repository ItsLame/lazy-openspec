package openspec

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestLiveClient exercises the real openspec CLI against whatever OpenSpec root
// the test runs under. It skips cleanly when the CLI or a root is absent, so it
// is safe in CI, but validates the full shell-out + decode path locally.
func TestLiveClient(t *testing.T) {
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := c.Changes(ctx)
	if errors.Is(err, ErrCLINotFound) || errors.Is(err, ErrNoRoot) {
		t.Skipf("skipping live test: %v", err)
	}
	if err != nil {
		t.Fatalf("Changes: %v", err)
	}
	if list.Root.Path == "" {
		t.Errorf("expected a resolved root path")
	}
	t.Logf("live: %d change(s) under %s", len(list.Changes), list.Root.Path)

	// If there is at least one change, its status and detail must decode too.
	if len(list.Changes) > 0 {
		name := list.Changes[0].Name
		if _, err := c.Status(ctx, name); err != nil {
			t.Errorf("Status(%s): %v", name, err)
		}
		if _, err := c.ChangeDetail(ctx, name); err != nil {
			t.Errorf("ChangeDetail(%s): %v", name, err)
		}
	}
}
