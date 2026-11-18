package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileGivesDefaultList(t *testing.T) {
	s, err := load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Lists) != 1 || s.Lists[0].Name != defaultListName || s.Current != 0 {
		t.Fatalf("got %+v, want one empty %q list", s, defaultListName)
	}
}

func TestLoadMigratesV010Format(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todos.json")
	old := `[
  {"text": "buy milk", "done": true},
  {"text": "fix bug", "done": false, "priority": 1}
]`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Lists) != 1 || s.Lists[0].Name != defaultListName {
		t.Fatalf("want a single %q list, got %+v", defaultListName, s.Lists)
	}
	items := s.Lists[0].Items
	if len(items) != 2 || items[0].Text != "buy milk" || !items[0].Done || items[1].Priority != prioHigh {
		t.Fatalf("items not migrated correctly: %+v", items)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "todos.json")
	want := store{Current: 1, Lists: []list{
		{Name: "Work", Items: []item{{Text: "ship it", Priority: prioMedium}}},
		{Name: "Home"},
	}}
	if err := save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Current != 1 || len(got.Lists) != 2 || got.Lists[0].Name != "Work" ||
		got.Lists[0].Items[0] != want.Lists[0].Items[0] || got.Lists[1].Name != "Home" {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestLoadClampsCurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todos.json")
	if err := os.WriteFile(path, []byte(`{"current": 7, "lists": [{"name": "A", "items": []}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Current != 0 {
		t.Fatalf("current = %d, want 0", s.Current)
	}
}
