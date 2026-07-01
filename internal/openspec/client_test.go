package openspec

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeChangeList(t *testing.T) {
	raw := []byte(`{"changes":[{"name":"add-x","completedTasks":2,"totalTasks":5,"status":"in-progress"}],"root":{"path":"/tmp","source":"nearest"}}`)
	cl, err := decode[ChangeList](raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(cl.Changes) != 1 || cl.Changes[0].Name != "add-x" {
		t.Fatalf("unexpected: %+v", cl)
	}
	if cl.Changes[0].Percent() != 40 {
		t.Errorf("percent = %d, want 40", cl.Changes[0].Percent())
	}
	if cl.Changes[0].Lifecycle() != Active {
		t.Errorf("lifecycle = %v, want Active", cl.Changes[0].Lifecycle())
	}
}

func TestDecodeTrimsWarningPreamble(t *testing.T) {
	raw := []byte("Warning: Ignoring flags not applicable to change: scenarios\n{\"id\":\"c\",\"deltaCount\":1,\"deltas\":[]}")
	cd, err := decode[ChangeDetail](raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if cd.ID != "c" || cd.DeltaCount != 1 {
		t.Fatalf("unexpected: %+v", cd)
	}
}

func TestLifecycleBuckets(t *testing.T) {
	cases := []struct {
		c    ChangeSummary
		want Lifecycle
	}{
		{ChangeSummary{TotalTasks: 0}, Draft},
		{ChangeSummary{TotalTasks: 3, CompletedTasks: 0}, Active},
		{ChangeSummary{TotalTasks: 3, CompletedTasks: 3}, Completed},
	}
	for _, tc := range cases {
		if got := tc.c.Lifecycle(); got != tc.want {
			t.Errorf("%+v => %v, want %v", tc.c, got, tc.want)
		}
	}
}

func TestClassifyMapsNoRoot(t *testing.T) {
	err := classify(errors.New("exit 1"), "No openspec directory found")
	if !errors.Is(err, ErrNoRoot) {
		t.Errorf("want ErrNoRoot, got %v", err)
	}
}

func TestStoreArgs(t *testing.T) {
	if args := New().storeArgs(); args != nil {
		t.Errorf("no store => nil args, got %v", args)
	}
	got := New(WithStore("acme")).storeArgs()
	if strings.Join(got, " ") != "--store acme" {
		t.Errorf("store args = %v", got)
	}
}
