package main

import (
	"path/filepath"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSortByPriority(t *testing.T) {
	items := []item{
		{Text: "none-a"},
		{Text: "low", Priority: prioLow},
		{Text: "done-high", Done: true, Priority: prioHigh},
		{Text: "high-a", Priority: prioHigh},
		{Text: "done-none", Done: true},
		{Text: "medium", Priority: prioMedium},
		{Text: "high-b", Priority: prioHigh},
		{Text: "none-b"},
	}
	got, cursor := sortByPriority(items, 1) // cursor on "low"

	want := []string{"high-a", "high-b", "medium", "low", "none-a", "none-b", "done-high", "done-none"}
	for i, w := range want {
		if got[i].Text != w {
			t.Fatalf("position %d: got %q, want %q (full: %v)", i, got[i].Text, w, got)
		}
	}
	if got[cursor].Text != "low" {
		t.Errorf("cursor should follow %q, landed on %q", "low", got[cursor].Text)
	}
	if items[0].Text != "none-a" {
		t.Error("input slice was modified")
	}
}

// press feeds keys to the model. Single characters are typed as text;
// "enter", "esc", "tab" and "shift+tab" are sent as special keys.
func press(t *testing.T, m model, keys ...string) model {
	t.Helper()
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab}
		case "shift+tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
		default:
			for _, r := range k {
				next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
				m = next.(model)
			}
			continue
		}
		next, _ := m.Update(msg)
		m = next.(model)
	}
	return m
}

func names(m model) []string {
	var out []string
	for _, l := range m.data.Lists {
		out = append(out, l.Name)
	}
	return out
}

func newTestModel(t *testing.T) model {
	t.Helper()
	path := filepath.Join(t.TempDir(), "todos.json")
	s, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	m := newModel(path, s)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	return next.(model)
}

func TestListLifecycle(t *testing.T) {
	m := newTestModel(t)

	// Create two lists; each new list is inserted after the current one and selected.
	m = press(t, m, "n", "Work", "enter", "n", "Home", "enter")
	if got := names(m); !slices.Equal(got, []string{"Todo", "Work", "Home"}) || m.data.Current != 2 {
		t.Fatalf("after creating: lists %v, current %d", got, m.data.Current)
	}

	// Items are per list.
	m = press(t, m, "a", "paint fence", "enter")
	m = press(t, m, "tab") // wraps to Todo
	if m.data.Current != 0 || len(m.cur().Items) != 0 {
		t.Fatalf("tab should wrap to empty Todo list, got current %d with %d items", m.data.Current, len(m.cur().Items))
	}
	m = press(t, m, "shift+tab")
	if m.cur().Name != "Home" || m.cur().Items[0].Text != "paint fence" {
		t.Fatalf("shift+tab should go back to Home with its item, got %q %+v", m.cur().Name, m.cur().Items)
	}

	// Rename pre-fills the old name; esc cancels.
	m = press(t, m, "r", "!", "esc")
	if m.cur().Name != "Home" {
		t.Fatalf("esc should cancel rename, name is %q", m.cur().Name)
	}
	m = press(t, m, "r", "s", "enter")
	if m.cur().Name != "Homes" {
		t.Fatalf("rename should append to existing name, got %q", m.cur().Name)
	}

	// Delete needs confirmation.
	m = press(t, m, "D", "n")
	if len(m.data.Lists) != 3 {
		t.Fatal("answering n should not delete the list")
	}
	m = press(t, m, "D", "y")
	if got := names(m); !slices.Equal(got, []string{"Todo", "Work"}) || m.data.Current != 1 {
		t.Fatalf("after delete: lists %v, current %d", got, m.data.Current)
	}

	// Everything was saved, including which list is current.
	saved, err := load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(model{data: saved}); !slices.Equal(got, []string{"Todo", "Work"}) || saved.Current != 1 {
		t.Fatalf("saved: lists %v, current %d", got, saved.Current)
	}
}

func TestCannotDeleteOnlyList(t *testing.T) {
	m := press(t, newTestModel(t), "D", "y")
	if len(m.data.Lists) != 1 || m.mode != modeNormal {
		t.Fatalf("only list was deleted or mode stuck: %d lists, mode %d", len(m.data.Lists), m.mode)
	}
}

func TestToggleCompletion(t *testing.T) {
	m := press(t, newTestModel(t), "a", "walk dog", "enter")
	m = press(t, m, " ")
	if !m.cur().Items[0].Done {
		t.Fatal("space should mark the item done")
	}
	m = press(t, m, " ")
	if m.cur().Items[0].Done {
		t.Fatal("space on a done item should mark it not done")
	}
	m = press(t, m, "x", "x")
	if m.cur().Items[0].Done {
		t.Fatal("x twice should leave the item not done")
	}
	saved, err := load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Lists[0].Items[0].Done {
		t.Fatal("un-completing should be saved")
	}
}
