package openspec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Common typed errors so callers (and the UI) can render actionable messages
// instead of raw failures.
var (
	// ErrCLINotFound means the `openspec` executable is not on PATH.
	ErrCLINotFound = errors.New("openspec CLI not found on PATH")
	// ErrNoRoot means no openspec/ root could be resolved.
	ErrNoRoot = errors.New("no OpenSpec project found (no openspec/ root)")
)

// Client runs the openspec CLI and decodes its JSON output. It is safe for
// concurrent use and memoizes results until Invalidate is called.
type Client struct {
	bin   string // resolved executable name/path
	store string // optional --store <id>, applied to commands that support it

	mu    sync.Mutex
	cache map[string][]byte // key -> raw stdout
}

// Option configures a Client.
type Option func(*Client)

// WithStore targets a registered store, threading `--store <id>` through the
// commands that accept it.
func WithStore(id string) Option {
	return func(c *Client) { c.store = id }
}

// WithBinary overrides the executable name (mainly for tests).
func WithBinary(bin string) Option {
	return func(c *Client) { c.bin = bin }
}

// New builds a Client. It does not touch the filesystem or network.
func New(opts ...Option) *Client {
	c := &Client{bin: "openspec", cache: map[string][]byte{}}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Store returns the configured store id (empty when none).
func (c *Client) Store() string { return c.store }

// Invalidate clears the whole cache so the next call re-runs the CLI.
func (c *Client) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = map[string][]byte{}
}

// storeArgs returns the --store flag pair when a store is set. Only pass these
// to commands that actually accept --store.
func (c *Client) storeArgs() []string {
	if c.store == "" {
		return nil
	}
	return []string{"--store", c.store}
}

// run executes `openspec <args...>` and returns stdout. Results are cached by
// the full argument vector. Classifiable failures are mapped to typed errors.
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	key := strings.Join(args, "\x00")

	c.mu.Lock()
	if out, ok := c.cache[key]; ok {
		c.mu.Unlock()
		return out, nil
	}
	c.mu.Unlock()

	if _, err := exec.LookPath(c.bin); err != nil {
		return nil, ErrCLINotFound
	}

	cmd := exec.CommandContext(ctx, c.bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, classify(err, stderr.String())
	}

	out := stdout.Bytes()
	c.mu.Lock()
	c.cache[key] = out
	c.mu.Unlock()
	return out, nil
}

// classify turns an exec failure into a typed or annotated error.
func classify(err error, stderr string) error {
	s := strings.ToLower(stderr)
	switch {
	case strings.Contains(s, "no openspec") || strings.Contains(s, "openspec directory"):
		return ErrNoRoot
	case strings.TrimSpace(stderr) != "":
		return fmt.Errorf("openspec: %s", strings.TrimSpace(stderr))
	default:
		return err
	}
}

// decode unmarshals tolerantly: unknown fields are ignored (default json
// behaviour), so additive CLI changes do not break the UI.
func decode[T any](raw []byte) (T, error) {
	var v T
	// The CLI occasionally prints a leading "Warning:" line before JSON; trim to
	// the first JSON delimiter so decoding stays robust.
	raw = trimToJSON(raw)
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, fmt.Errorf("decoding openspec output: %w", err)
	}
	return v, nil
}

// trimToJSON drops any non-JSON preamble (e.g. a warning line) before the first
// '{' or '['.
func trimToJSON(raw []byte) []byte {
	for i, b := range raw {
		if b == '{' || b == '[' {
			return raw[i:]
		}
	}
	return raw
}

// DefaultTimeout bounds a single CLI invocation.
const DefaultTimeout = 15 * time.Second

// Changes returns the change list (`openspec list --json`).
func (c *Client) Changes(ctx context.Context) (ChangeList, error) {
	args := append([]string{"list", "--json"}, c.storeArgs()...)
	raw, err := c.run(ctx, args...)
	if err != nil {
		return ChangeList{}, err
	}
	return decode[ChangeList](raw)
}

// Specs returns the spec list (`openspec list --specs --json`).
func (c *Client) Specs(ctx context.Context) (SpecList, error) {
	args := append([]string{"list", "--specs", "--json"}, c.storeArgs()...)
	raw, err := c.run(ctx, args...)
	if err != nil {
		return SpecList{}, err
	}
	return decode[SpecList](raw)
}

// Status returns artifact/task status for a change.
func (c *Client) Status(ctx context.Context, change string) (Status, error) {
	args := append([]string{"status", "--change", change, "--json"}, c.storeArgs()...)
	raw, err := c.run(ctx, args...)
	if err != nil {
		return Status{}, err
	}
	return decode[Status](raw)
}

// ChangeDetail returns the deltas for a change (`openspec show <name> --json`).
func (c *Client) ChangeDetail(ctx context.Context, change string) (ChangeDetail, error) {
	args := append([]string{"show", change, "--json", "--type", "change"}, c.storeArgs()...)
	raw, err := c.run(ctx, args...)
	if err != nil {
		return ChangeDetail{}, err
	}
	return decode[ChangeDetail](raw)
}

// SpecDetail returns requirements for a spec (`openspec spec show <id> --json`).
func (c *Client) SpecDetail(ctx context.Context, id string) (SpecDetail, error) {
	raw, err := c.run(ctx, "spec", "show", id, "--json")
	if err != nil {
		return SpecDetail{}, err
	}
	return decode[SpecDetail](raw)
}

// ArtifactMarkdown returns the raw markdown for one of a change's artifacts by
// reading it from the change directory reported by Status. The UI renders this
// through Glamour. Returns ("", nil) when the artifact does not exist yet.
func (c *Client) ArtifactMarkdown(ctx context.Context, change, artifactPath string) (string, error) {
	// Kept as a thin helper; callers pass the resolved path from Status.
	return readIfExists(artifactPath)
}
