package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type priority int

const (
	prioNone priority = iota
	prioHigh
	prioMedium
	prioLow
)

type item struct {
	Text     string   `json:"text"`
	Done     bool     `json:"done"`
	Priority priority `json:"priority,omitempty"`
}

type list struct {
	Name  string `json:"name"`
	Items []item `json:"items"`

	// View state, kept per list while the app runs but not saved.
	cursor int
	top    int
}

// store is everything saved to disk.
type store struct {
	Current int    `json:"current"` // index of the list shown on startup
	Lists   []list `json:"lists"`
}

const defaultListName = "Todo"

// dataFile returns where the todo lists are stored: $TODOUI_FILE if set,
// otherwise $XDG_DATA_HOME/todoui/todos.json (default ~/.local/share).
func dataFile() (string, error) {
	if p := os.Getenv("TODOUI_FILE"); p != "" {
		return p, nil
	}
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "todoui", "todos.json"), nil
}

// load reads the store, accepting both the current format and the v0.1.0
// format (a bare JSON array of items), which becomes a single list.
// It always returns at least one list.
func load(path string) (store, error) {
	var s store
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// First run: start with an empty default list.
	case err != nil:
		return s, err
	case bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")):
		var items []item
		if err := json.Unmarshal(data, &items); err != nil {
			return s, err
		}
		s.Lists = []list{{Name: defaultListName, Items: items}}
	default:
		if err := json.Unmarshal(data, &s); err != nil {
			return s, err
		}
	}
	if len(s.Lists) == 0 {
		s.Lists = []list{{Name: defaultListName}}
	}
	s.Current = max(0, min(s.Current, len(s.Lists)-1))
	return s, nil
}

// save writes atomically so a crash mid-write can't corrupt the lists.
func save(path string, s store) error {
	for i := range s.Lists {
		if s.Lists[i].Items == nil {
			s.Lists[i].Items = []item{}
		}
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
