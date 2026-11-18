// todoui is a simple todo TUI.
package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// version is set at release time by GoReleaser via -ldflags.
var version = "dev"

const help = "a:add  space:complete  J/K:move  d:delete  j/k:nav  q:quit"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	doneStyle     = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

type model struct {
	path   string
	items  []item
	cursor int
	top    int // index of the first visible item
	width  int
	height int
	adding bool
	input  textinput.Model
	err    error
}

func newModel(path string, items []item) model {
	ti := textinput.New()
	ti.Prompt = "New item: "
	return model{path: path, items: items, input: ti}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(1, msg.Width-len(m.input.Prompt)-1))
		m.scroll()
		return m, nil
	case tea.KeyPressMsg:
		if m.adding {
			return m.updateAdding(msg)
		}
		return m.updateList(msg)
	}
	if m.adding {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) updateAdding(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if text := strings.TrimSpace(m.input.Value()); text != "" {
			// Insert below the cursor (or at the top of an empty list).
			at := 0
			if len(m.items) > 0 {
				at = m.cursor + 1
			}
			m.items = append(m.items[:at], append([]item{{Text: text}}, m.items[at:]...)...)
			m.cursor = at
			m.persist()
		}
		fallthrough
	case "esc":
		m.adding = false
		m.input.Blur()
		m.input.Reset()
		m.scroll()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	n := len(m.items)
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "a":
		m.adding = true
		return m, m.input.Focus()
	case "j", "down":
		m.cursor = min(m.cursor+1, max(n-1, 0))
	case "k", "up":
		m.cursor = max(m.cursor-1, 0)
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = max(n-1, 0)
	case "space", "x", "enter":
		if n > 0 {
			m.items[m.cursor].Done = true
			m.persist()
		}
	case "J", "shift+down":
		if m.cursor < n-1 {
			m.items[m.cursor], m.items[m.cursor+1] = m.items[m.cursor+1], m.items[m.cursor]
			m.cursor++
			m.persist()
		}
	case "K", "shift+up":
		if m.cursor > 0 && n > 0 {
			m.items[m.cursor], m.items[m.cursor-1] = m.items[m.cursor-1], m.items[m.cursor]
			m.cursor--
			m.persist()
		}
	case "d":
		if n > 0 {
			m.items = append(m.items[:m.cursor], m.items[m.cursor+1:]...)
			m.cursor = min(m.cursor, max(len(m.items)-1, 0))
			m.persist()
		}
	}
	m.scroll()
	return m, nil
}

func (m *model) persist() {
	m.err = save(m.path, m.items)
}

// listHeight is the number of rows available for items
// (title + blank line above, footer below).
func (m model) listHeight() int { return max(1, m.height-3) }

// scroll keeps the cursor inside the visible window.
func (m *model) scroll() {
	h := m.listHeight()
	if m.cursor < m.top {
		m.top = m.cursor
	} else if m.cursor >= m.top+h {
		m.top = m.cursor - h + 1
	}
	m.top = max(0, min(m.top, max(len(m.items)-h, 0)))
}

func (m model) View() tea.View {
	var b strings.Builder
	done := 0
	for _, it := range m.items {
		if it.Done {
			done++
		}
	}
	fit := lipgloss.NewStyle().MaxWidth(max(m.width, 1))
	b.WriteString(titleStyle.Render(fmt.Sprintf(" todoui  (%d/%d done)", done, len(m.items))) + "\n\n")

	rows := 0
	if len(m.items) == 0 {
		b.WriteString(dimStyle.Render("  No items. Press 'a' to add one.") + "\n")
		rows = 1
	}
	for i := m.top; i < len(m.items) && rows < m.listHeight(); i, rows = i+1, rows+1 {
		it := m.items[i]
		mark, text := "[ ]", it.Text
		if it.Done {
			mark, text = "[x]", doneStyle.Render(text)
		}
		line := fmt.Sprintf(" %s %s ", mark, text)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		b.WriteString(fit.Render(" "+line) + "\n")
	}
	for ; rows < m.listHeight(); rows++ {
		b.WriteString("\n")
	}

	switch {
	case m.adding:
		b.WriteString(m.input.View())
	case m.err != nil:
		b.WriteString(fit.Render(errStyle.Render("save failed: " + m.err.Error())))
	default:
		b.WriteString(fit.Render(dimStyle.Render(help)))
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
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
	items, err := load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "todoui: reading %s: %v\n", path, err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(path, items)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "todoui:", err)
		os.Exit(1)
	}
}
