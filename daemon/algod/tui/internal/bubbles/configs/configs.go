package configs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/node"
)

const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type Model struct {
	heightMargin int

	ready    bool
	viewport viewport.Model
	server   *algod.Server
}

func New(s *algod.Server, heightMargin int) Model {
	m := Model{
		viewport:     viewport.New(0, 0),
		heightMargin: heightMargin,
		server:       s,
	}
	m.setSize(80, 20)
	return m
}

type ConfigContent string

func (m Model) getContent() tea.Cmd {
	return func() tea.Msg {
		return ConfigContent(node.GetConfigs(algod.GetNode(m.server)))
	}
}

func (m Model) Init() tea.Cmd {
	return m.getContent()
}

func (m *Model) setSize(width, height int) {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())

	m.viewport.Width = 70
	m.viewport.Height = height - m.heightMargin - headerHeight - footerHeight
	//m.viewport.YPosition = headerHeight + 1
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	//
	switch msg := msg.(type) {
	case ConfigContent:
		m.ready = true
		//m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
		// this seems to work.
		//m.viewport.SetContent(strings.Repeat("This is a teat\n", 100))

		m.viewport.SetContent(strings.ReplaceAll(string(msg), "\t", "    "))

	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView()))
	return builder.String()
}

func (m Model) headerView() string {
	title := titleStyle.Render("Node configurations")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
