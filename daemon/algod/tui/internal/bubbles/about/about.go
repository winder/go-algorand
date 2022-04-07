package about

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	width        int
	height       int
	heightMargin int
}

func New(heightMargin int) Model {
	return Model{
		heightMargin: heightMargin,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	// -1 because the empty line at the end counts
	count := m.height - m.heightMargin - 1
	return strings.Repeat("About tab\n", count)
}
