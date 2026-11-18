package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type item struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// dataFile returns where the todo list is stored: $TODOUI_FILE if set,
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

func load(path string) ([]item, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var items []item
	return items, json.Unmarshal(data, &items)
}

// save writes atomically so a crash mid-write can't corrupt the list.
func save(path string, items []item) error {
	if items == nil {
		items = []item{}
	}
	data, err := json.MarshalIndent(items, "", "  ")
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
