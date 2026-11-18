package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	itemHelp = "a:add  space:toggle  1/2/3/0:priority  s:sort  J/K:move  d:delete  q:quit"
	listHelp = "tab/shift+tab:switch list  n:new list  r:rename list  D:delete list"
)

var (
	dimStyle    = lipgloss.NewStyle().Faint(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	activeTab   = lipgloss.NewStyle().Bold(true).Reverse(true)
	inactiveTab = lipgloss.NewStyle().Faint(true)

	// Text color for each priority; prioNone keeps the terminal default.
	priorityColors = map[priority]ansi.BasicColor{
		prioHigh:   ansi.BrightRed,
		prioMedium: ansi.BrightYellow,
		prioLow:    ansi.BrightBlue,
	}
)

// listHeight is the number of rows available for items
// (tab bar + blank line above, two footer lines below).
func (m model) listHeight() int { return max(1, m.height-4) }

func (m model) View() tea.View {
	var b strings.Builder
	fit := lipgloss.NewStyle().MaxWidth(max(m.width, 1))
	l := m.cur()

	b.WriteString(fit.Render(m.renderTabs()) + "\n\n")

	rows := 0
	if len(l.Items) == 0 {
		b.WriteString(dimStyle.Render("  No items. Press 'a' to add one.") + "\n")
		rows = 1
	}
	for i := l.top; i < len(l.Items) && rows < m.listHeight(); i, rows = i+1, rows+1 {
		b.WriteString(fit.Render(" "+renderItem(l.Items[i], i == l.cursor)) + "\n")
	}
	for ; rows < m.listHeight(); rows++ {
		b.WriteString("\n")
	}

	// Footer: two lines.
	var line1, line2 string
	switch m.mode {
	case modeAddItem, modeNewList, modeRenameList:
		line1 = dimStyle.Render("enter:save  esc:cancel")
		line2 = m.input.View()
	case modeConfirmDelete:
		line1 = errStyle.Render(fmt.Sprintf("Delete list %q and its %d item(s)?", l.Name, len(l.Items)))
		line2 = "y: delete  any other key: cancel"
	default:
		line1, line2 = dimStyle.Render(itemHelp), dimStyle.Render(listHelp)
		if m.status != "" {
			line1 = m.status
		}
	}
	if m.err != nil {
		line1 = errStyle.Render("save failed: " + m.err.Error())
	}
	b.WriteString(fit.Render(line1) + "\n" + fit.Render(line2))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// renderTabs draws one tab per list, plus the done count for the current
// list. If the tabs don't fit, leading tabs are dropped (marked with "‹")
// so the current list is always visible.
func (m model) renderTabs() string {
	tabs := make([]string, len(m.data.Lists))
	for i, l := range m.data.Lists {
		if i == m.data.Current {
			tabs[i] = activeTab.Render(" " + l.Name + " ")
		} else {
			tabs[i] = inactiveTab.Render(" " + l.Name + " ")
		}
	}
	l := m.cur()
	done := 0
	for _, it := range l.Items {
		if it.Done {
			done++
		}
	}
	count := dimStyle.Render(fmt.Sprintf("  %d/%d done", done, len(l.Items)))

	first := 0
	width := func() int {
		w := lipgloss.Width(strings.Join(tabs[first:m.data.Current+1], " ")) + lipgloss.Width(count)
		if first > 0 {
			w += 2 // "‹ "
		}
		return w
	}
	for m.width > 0 && first < m.data.Current && width() > m.width {
		first++
	}
	// The count goes right after the current tab so it stays visible;
	// tabs after it are simply cut off at the screen edge if too wide.
	bar := strings.Join(tabs[first:m.data.Current+1], " ")
	if first > 0 {
		bar = dimStyle.Render("‹ ") + bar
	}
	bar += count
	if rest := tabs[m.data.Current+1:]; len(rest) > 0 {
		bar += " " + strings.Join(rest, " ")
	}
	return bar
}

// renderItem builds one SGR style for the whole row. It uses x/ansi rather
// than lipgloss because lipgloss renders the spaces in struck-through text
// separately and drops the reverse-video highlight from them.
func renderItem(it item, selected bool) string {
	st := ansi.NewStyle()
	if selected {
		st = st.Reverse(true)
	}
	mark := "[ ]"
	text := st
	if c, ok := priorityColors[it.Priority]; ok {
		text = text.ForegroundColor(c)
	}
	if it.Done {
		mark = "[x]"
		text = text.Faint().Strikethrough(true)
	}
	return st.Styled(" "+mark+" ") + text.Styled(it.Text) + st.Styled(" ")
}
