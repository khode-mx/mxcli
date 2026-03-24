package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const jumperMaxShow = 12

// JumpToNodeMsg is emitted when the user selects a node from the jumper.
type JumpToNodeMsg struct {
	QName    string
	NodeType string
}

// JumperView is a fuzzy-search modal for jumping to any node in the project.
type JumperView struct {
	input   textinput.Model
	items   []PickerItem
	matches []pickerMatch
	cursor  int
	offset  int
	width   int
	height  int
}

// NewJumperView creates a JumperView populated with the given items.
func NewJumperView(items []PickerItem, width, height int) JumperView {
	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.Placeholder = "jump to..."
	ti.CharLimit = 200
	ti.Focus()

	jv := JumperView{
		input:  ti,
		items:  items,
		width:  width,
		height: height,
	}
	jv.filterItems()
	return jv
}

func (jv *JumperView) filterItems() {
	query := strings.TrimSpace(jv.input.Value())
	jv.matches = nil
	for _, it := range jv.items {
		if query == "" {
			jv.matches = append(jv.matches, pickerMatch{item: it})
			continue
		}
		if ok, sc := fuzzyScore(it.QName, query); ok {
			jv.matches = append(jv.matches, pickerMatch{item: it, score: sc})
		}
	}
	// Sort by score descending (insertion sort, small n)
	for i := 1; i < len(jv.matches); i++ {
		for j := i; j > 0 && jv.matches[j].score > jv.matches[j-1].score; j-- {
			jv.matches[j], jv.matches[j-1] = jv.matches[j-1], jv.matches[j]
		}
	}
	if jv.cursor >= len(jv.matches) {
		jv.cursor = max(0, len(jv.matches)-1)
	}
	jv.offset = 0
}

func (jv *JumperView) moveDown() {
	if len(jv.matches) == 0 {
		return
	}
	jv.cursor++
	if jv.cursor >= len(jv.matches) {
		jv.cursor = 0
		jv.offset = 0
	} else if jv.cursor >= jv.offset+jumperMaxShow {
		jv.offset = jv.cursor - jumperMaxShow + 1
	}
}

func (jv *JumperView) moveUp() {
	if len(jv.matches) == 0 {
		return
	}
	jv.cursor--
	if jv.cursor < 0 {
		jv.cursor = len(jv.matches) - 1
		jv.offset = max(0, jv.cursor-jumperMaxShow+1)
	} else if jv.cursor < jv.offset {
		jv.offset = jv.cursor
	}
}

func (jv JumperView) selected() PickerItem {
	if len(jv.matches) == 0 || jv.cursor >= len(jv.matches) {
		return PickerItem{}
	}
	return jv.matches[jv.cursor].item
}

// --- View interface ---

// Update handles key messages for the jumper modal.
func (jv JumperView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return jv, func() tea.Msg { return PopViewMsg{} }
		case "enter":
			sel := jv.selected()
			if sel.QName != "" {
				qname := sel.QName
				nodeType := sel.NodeType
				return jv, func() tea.Msg { return JumpToNodeMsg{QName: qname, NodeType: nodeType} }
			}
			return jv, func() tea.Msg { return PopViewMsg{} }
		case "up", "ctrl+p":
			jv.moveUp()
		case "down", "ctrl+n":
			jv.moveDown()
		default:
			var cmd tea.Cmd
			jv.input, cmd = jv.input.Update(msg)
			jv.filterItems()
			return jv, cmd
		}

	case tea.WindowSizeMsg:
		jv.width = msg.Width
		jv.height = msg.Height
	}
	return jv, nil
}

// Render draws the jumper as a centered modal box with an LLM anchor prefix.
func (jv JumperView) Render(width, height int) string {
	selSt := SelectedItemStyle
	normSt := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	dimSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	typeSt := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	boxWidth := max(30, min(60, width-10))

	// LLM anchor embedded at the top of the box content
	query := strings.TrimSpace(jv.input.Value())
	anchor := fmt.Sprintf("[mxcli:jump] > %s  %d matches", query, len(jv.matches))
	anchorStr := lipgloss.NewStyle().Foreground(MutedColor).Faint(true).Render(anchor)

	var sb strings.Builder
	sb.WriteString(anchorStr + "\n")
	sb.WriteString(jv.input.View() + "\n\n")

	end := min(jv.offset+jumperMaxShow, len(jv.matches))
	if jv.offset > 0 {
		sb.WriteString(dimSt.Render("  ↑ more") + "\n")
	}
	for i := jv.offset; i < end; i++ {
		it := jv.matches[i].item
		icon := IconFor(it.NodeType)
		label := icon + " " + it.QName
		typeLabel := it.NodeType
		if i == jv.cursor {
			sb.WriteString(selSt.Render(" "+label) + " " + typeSt.Render(typeLabel) + "\n")
		} else {
			sb.WriteString(normSt.Render(" "+label) + " " + dimSt.Render(typeLabel) + "\n")
		}
	}
	if end < len(jv.matches) {
		sb.WriteString(dimSt.Render("  ↓ more") + "\n")
	}
	sb.WriteString("\n" + dimSt.Render(fmt.Sprintf("  %d/%d", len(jv.matches), len(jv.items))))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(boxWidth).
		Render(sb.String())

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("0")))
}

// Hints returns jumper-specific key hints.
func (jv JumperView) Hints() []Hint {
	return []Hint{
		{Key: "↑/↓", Label: "navigate"},
		{Key: "Enter", Label: "jump"},
		{Key: "Esc", Label: "cancel"},
	}
}

// StatusInfo returns match count information.
func (jv JumperView) StatusInfo() StatusInfo {
	return StatusInfo{
		Mode:     "Jump",
		Position: fmt.Sprintf("%d/%d", len(jv.matches), len(jv.items)),
	}
}

// Mode returns ModeJumper.
func (jv JumperView) Mode() ViewMode {
	return ModeJumper
}
