package tui

import "github.com/charmbracelet/lipgloss"

// Semantic color tokens — AdaptiveColor picks Light/Dark based on terminal background.
var (
	FocusColor   = lipgloss.AdaptiveColor{Light: "62", Dark: "63"}
	AccentColor  = lipgloss.AdaptiveColor{Light: "214", Dark: "214"}
	MutedColor   = lipgloss.AdaptiveColor{Light: "245", Dark: "243"}
	AddedColor   = lipgloss.AdaptiveColor{Light: "28", Dark: "114"}
	RemovedColor = lipgloss.AdaptiveColor{Light: "124", Dark: "210"}
)

var (
	// Column separator: dim vertical bar between panels.
	SeparatorChar  = "│"
	SeparatorStyle = lipgloss.NewStyle().Foreground(MutedColor)

	// Tabs
	ActiveTabStyle   = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(FocusColor)
	InactiveTabStyle = lipgloss.NewStyle().Foreground(MutedColor)

	// Column title (e.g. "Entities", "Attributes")
	ColumnTitleStyle = lipgloss.NewStyle().Bold(true)

	// List items
	SelectedItemStyle = lipgloss.NewStyle().Reverse(true)
	DirectoryStyle    = lipgloss.NewStyle().Bold(true)
	LeafStyle         = lipgloss.NewStyle()

	// Breadcrumb
	BreadcrumbDimStyle     = lipgloss.NewStyle().Foreground(MutedColor)
	BreadcrumbCurrentStyle = lipgloss.NewStyle()

	// Loading / status
	LoadingStyle  = lipgloss.NewStyle().Italic(true).Foreground(MutedColor)
	PositionStyle = lipgloss.NewStyle().Foreground(MutedColor)

	// Preview mode label (MDL / NDSL toggle)
	PreviewModeStyle = lipgloss.NewStyle().Bold(true)

	// Hint bar: key name bold, description dim
	HintKeyStyle   = lipgloss.NewStyle().Bold(true)
	HintLabelStyle = lipgloss.NewStyle().Foreground(MutedColor)

	// Status bar (bottom line)
	StatusBarStyle = lipgloss.NewStyle().Foreground(MutedColor)

	// Command bar
	CmdBarStyle = lipgloss.NewStyle().Bold(true)

	// Focus indicator styles (Phase 2)
	FocusedTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(FocusColor)
	FocusedEdgeChar   = "▎"
	FocusedEdgeStyle  = lipgloss.NewStyle().Foreground(FocusColor)
	AccentStyle       = lipgloss.NewStyle().Foreground(AccentColor)
)
