package tui

import (
	"errors"
	"io/fs"
	"os"
	"sort"
)

// readFile returns a file's contents, or "" (no error) when it does not exist.
func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// writeFile writes content to path, creating/truncating the file.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

// listDirs returns the sorted names of subdirectories in dir (empty when the
// directory is absent).
func listDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
