package explorer

import (
	"github.com/algorand/go-algorand/node"
	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
)

type state int

const (
	blockState = iota
	paysetState
	txnState
)

type Model struct {
	state state

	width        int
	widthMargin  int
	height       int
	heightMargin int
	style        *style.Styles

	blockPerPage uint

	node *node.AlgorandFullNode

	blocks blocks

	table table.Model
}

func NewModel(server *algod.Server, styles *style.Styles, width, widthMargin, height, heightMargin int) Model {
	node := algod.GetNode(server)

	m := Model{
		state:        blockState,
		style:        styles,
		node:         node,
		width:        width,
		widthMargin:  widthMargin,
		height:       height,
		heightMargin: heightMargin,
	}
	m.initBlock()
	return m
}

func (m Model) Init() tea.Cmd {
	// Default page.
	return m.InitBlock()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.state {
	case blockState:
		return m.UpdateBlock(msg)
	case paysetState:
		return m, nil
	case txnState:
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	return m.ViewBlock()
	switch m.state {
	case blockState:
		return m.ViewBlock()
	case paysetState:
		return ""
	case txnState:
		return ""
	}
	return ""
}
