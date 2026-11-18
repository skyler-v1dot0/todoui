// todoui is a simple todo TUI.
package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// version is set at release time by GoReleaser via -ldflags.
var version = "dev"

// mode is what the keyboard is currently driving.
type mode int

const (
	modeNormal        mode = iota
	modeAddItem            // typing a new item
	modeNewList            // typing a new list's name
	modeRenameList         // editing the current list's name
	modeConfirmDelete      // waiting for y/n to delete the current list
)

type model struct {
	path   string
	data   store
	mode   mode
	input  textinput.Model
	width  int
	height int
	status string // one-off message shown in the footer until the next key
	err    error
}

func newModel(path string, data store) model {
	return model{path: path, data: data, input: textinput.New()}
}

// cur returns the list being shown.
func (m *model) cur() *list { return &m.data.Lists[m.data.Current] }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.sizeInput()
		m.scroll()
		return m, nil
	case tea.KeyPressMsg:
		m.status = ""
		switch m.mode {
		case modeNormal:
			return m.updateNormal(msg)
		case modeConfirmDelete:
			return m.updateConfirmDelete(msg)
		default:
			return m.updatePrompt(msg)
		}
	}
	if m.mode != modeNormal {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// startPrompt switches to a text-entry mode with the given prompt and text.
func (m *model) startPrompt(md mode, prompt, value string) tea.Cmd {
	m.mode = md
	m.input.Prompt = prompt
	m.input.SetValue(value)
	m.input.CursorEnd()
	m.sizeInput()
	return m.input.Focus()
}

func (m *model) sizeInput() {
	m.input.SetWidth(max(1, m.width-len(m.input.Prompt)-1))
}

func (m model) updatePrompt(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if text := strings.TrimSpace(m.input.Value()); text != "" {
			m.submitPrompt(text)
		}
		fallthrough
	case "esc":
		m.mode = modeNormal
		m.input.Blur()
		m.input.Reset()
		m.scroll()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) submitPrompt(text string) {
	switch m.mode {
	case modeAddItem:
		// Insert below the cursor (or at the top of an empty list).
		l := m.cur()
		at := 0
		if len(l.Items) > 0 {
			at = l.cursor + 1
		}
		l.Items = slices.Insert(l.Items, at, item{Text: text})
		l.cursor = at
	case modeNewList:
		at := m.data.Current + 1
		m.data.Lists = slices.Insert(m.data.Lists, at, list{Name: text})
		m.data.Current = at
	case modeRenameList:
		m.cur().Name = text
	}
	m.persist()
}

func (m model) updateConfirmDelete(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.mode = modeNormal
	if msg.String() != "y" && msg.String() != "Y" {
		m.status = "Delete cancelled."
		return m, nil
	}
	name := m.cur().Name
	m.data.Lists = slices.Delete(m.data.Lists, m.data.Current, m.data.Current+1)
	m.data.Current = min(m.data.Current, len(m.data.Lists)-1)
	m.status = fmt.Sprintf("Deleted list %q.", name)
	m.persist()
	return m, nil
}

func (m model) updateNormal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	l := m.cur()
	n := len(l.Items)
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	// Lists
	case "tab", "l", "right":
		m.switchList(1)
	case "shift+tab", "h", "left":
		m.switchList(-1)
	case "n":
		return m, m.startPrompt(modeNewList, "New list: ", "")
	case "r":
		return m, m.startPrompt(modeRenameList, "Rename list: ", l.Name)
	case "D":
		if len(m.data.Lists) == 1 {
			m.status = "Can't delete the only list."
		} else {
			m.mode = modeConfirmDelete
		}

	// Items
	case "a":
		return m, m.startPrompt(modeAddItem, "New item: ", "")
	case "j", "down":
		l.cursor = min(l.cursor+1, max(n-1, 0))
	case "k", "up":
		l.cursor = max(l.cursor-1, 0)
	case "g", "home":
		l.cursor = 0
	case "G", "end":
		l.cursor = max(n-1, 0)
	case "space", "x", "enter":
		if n > 0 {
			l.Items[l.cursor].Done = !l.Items[l.cursor].Done
			m.persist()
		}
	case "1", "2", "3", "0":
		if n > 0 {
			l.Items[l.cursor].Priority = priority(msg.String()[0] - '0')
			m.persist()
		}
	case "s":
		if n > 1 {
			l.Items, l.cursor = sortByPriority(l.Items, l.cursor)
			m.persist()
		}
	case "J", "shift+down":
		if l.cursor < n-1 {
			l.Items[l.cursor], l.Items[l.cursor+1] = l.Items[l.cursor+1], l.Items[l.cursor]
			l.cursor++
			m.persist()
		}
	case "K", "shift+up":
		if l.cursor > 0 && n > 0 {
			l.Items[l.cursor], l.Items[l.cursor-1] = l.Items[l.cursor-1], l.Items[l.cursor]
			l.cursor--
			m.persist()
		}
	case "d":
		if n > 0 {
			l.Items = slices.Delete(l.Items, l.cursor, l.cursor+1)
			l.cursor = min(l.cursor, max(len(l.Items)-1, 0))
			m.persist()
		}
	}
	m.scroll()
	return m, nil
}

// switchList moves to the next (+1) or previous (-1) list, wrapping around.
func (m *model) switchList(delta int) {
	n := len(m.data.Lists)
	if n < 2 {
		return
	}
	m.data.Current = (m.data.Current + delta + n) % n
	m.persist() // remember the current list for next launch
}

// sortRank orders items: open before done, then high → medium → low → none.
func sortRank(it item) int {
	rank := int(it.Priority)
	if it.Priority == prioNone {
		rank = int(prioLow) + 1
	}
	if it.Done {
		rank += 10
	}
	return rank
}

// sortByPriority returns the items stably sorted by sortRank (so manual order
// is kept within each group) and the new index of the item at cursor.
func sortByPriority(items []item, cursor int) ([]item, int) {
	order := make([]int, len(items))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return cmp.Compare(sortRank(items[a]), sortRank(items[b]))
	})
	sorted := make([]item, len(items))
	newCursor := cursor
	for to, from := range order {
		sorted[to] = items[from]
		if from == cursor {
			newCursor = to
		}
	}
	return sorted, newCursor
}

func (m *model) persist() {
	m.err = save(m.path, m.data)
}

// scroll keeps the current list's cursor inside the visible window.
func (m *model) scroll() {
	l := m.cur()
	h := m.listHeight()
	if l.cursor < l.top {
		l.top = l.cursor
	} else if l.cursor >= l.top+h {
		l.top = l.cursor - h + 1
	}
	l.top = max(0, min(l.top, max(len(l.Items)-h, 0)))
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("todoui", version)
		return
	}
	path, err := dataFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, "todoui:", err)
		os.Exit(1)
	}
	data, err := load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "todoui: reading %s: %v\n", path, err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(path, data)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "todoui:", err)
		os.Exit(1)
	}
}
