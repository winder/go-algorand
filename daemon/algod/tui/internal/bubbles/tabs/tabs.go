package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	activeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#527772")).
			Foreground(lipgloss.Color("#6dd588"))
	inactiveStyle = lipgloss.NewStyle()
)

type Model struct {
	width int

	index int
	tabs  []string

	tabWidth int

	ActiveStyle   lipgloss.Style
	InactiveStyle lipgloss.Style
}

func New(tabs []string) Model {
	max := 0
	for _, t := range tabs {
		if len(t) > max {
			max = len(t)
		}
	}
	return Model{
		width:         80,
		tabs:          tabs,
		tabWidth:      max,
		ActiveStyle:   activeStyle,
		InactiveStyle: inactiveStyle,
	}
}

func (m *Model) SetActiveIndex(i int) {
	m.index = i
}

func (m Model) GetActiveIndex() int {
	return m.index
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	return m, nil
}

func (m Model) View() string {
	var buf strings.Builder
	buf.WriteString("___")
	len := 3
	for i, col := range m.tabs {
		renderer := m.InactiveStyle
		if i == m.index {
			renderer = m.ActiveStyle
		}
		buf.WriteString(renderer.Render(strings.ReplaceAll(fmt.Sprintf("__%-*s__", m.tabWidth, col), " ", "_")))
		len += 4 + m.tabWidth
	}

	// TODO: If width-l is less than 0 this needs to be truncated instead
	buf.WriteString(strings.Repeat("_", m.width-len))
	return buf.String()
}
