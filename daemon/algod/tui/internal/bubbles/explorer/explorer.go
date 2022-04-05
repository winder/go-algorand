package explorer

import (
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
	// height
	// width
	// rowsPerPage
	// maxRound

	state state

	blocks blockModel
}

func NewModel(server *algod.Server, styles *style.Styles, width, widthMargin, height, heightMargin int) Model {
	node := algod.GetNode(server)
	return Model{
		state:  blockState,
		blocks: newBlockModel(node, styles, width, widthMargin, height, heightMargin),
	}
}

func (m Model) Init() tea.Cmd {
	// Default page.
	return m.blocks.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch m.state {
	case blockState:
		blockModel, cmd := m.blocks.Update(msg)
		m.blocks = *blockModel
		return m, cmd
	case paysetState:
		return m, nil
	case txnState:
		return m, nil
	}

	blocks, blocksCommand := m.blocks.Update(msg)
	m.blocks = *blocks

	return m, tea.Batch(blocksCommand)
}

func (m Model) View() string {
	return m.blocks.View()
}
