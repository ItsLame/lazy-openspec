package openspec

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// readIfExists returns the file's contents, or ("", nil) when it does not
// exist. Any other error (permissions, etc.) is returned. Used to render
// artifact prose (proposal.md, design.md, tasks.md) that the CLI exposes no
// JSON form for.
func readIfExists(path string) (string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ArtifactPath joins a change root with an artifact's relative output path.
func ArtifactPath(changeRoot, rel string) string {
	return filepath.Join(changeRoot, rel)
}
