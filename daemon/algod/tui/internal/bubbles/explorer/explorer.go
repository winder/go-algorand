package explorer

import (
	"time"

	table "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/constants"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/node"
)

type state int

const (
	blockState = iota
	paysetState
	txnState
)

type blocks []blockItem
type txnItems []transactionItem

type Model struct {
	state state

	width        int
	widthMargin  int
	height       int
	heightMargin int
	style        *style.Styles

	blockPerPage uint

	node *node.AlgorandFullNode

	// for blocks page
	blocks blocks

	// cache for transactions page
	transactions txnItems

	// cache for txn page
	txn transactions.SignedTxnInBlock

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
		heightMargin: heightMargin + 6,
	}
	m.initBlocks()
	return m
}

type BlocksMsg struct {
	blocks []blockItem
	err    error
}

func (m Model) InitBlock() tea.Cmd {
	return m.getLatestBlockHeaders
}

func (m *Model) getLatestBlockHeaders() tea.Msg {
	// TODO: Only fetch if needed, check current latest vs actual latest
	var result BlocksMsg

	ledger := m.node.Ledger()
	latest := ledger.Latest()
	for m.blockPerPage > uint(len(result.blocks)) && latest > 0 {
		block, err := ledger.Block(latest)
		if err != nil {
			result.err = err
			return result
		}
		latest -= 1

		result.blocks = append(result.blocks, blockItem{&block})
	}
	return result
}

func (m Model) Init() tea.Cmd {
	// Default page.
	return m.getLatestBlockHeaders
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.table.SetSize(width-m.widthMargin, height-m.heightMargin)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var updateCmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// navigate into explorer views
		switch {
		case key.Matches(msg, constants.Keys.Forward):
			switch m.state {
			case blockState:
				m.state = paysetState
				switch block := m.table.SelectedRow().(type) {
				case blockItem:
					m.transactions = make([]transactionItem, 0)
					for _, txn := range block.Payset {
						t := txn
						m.transactions = append(m.transactions, transactionItem{&t})
					}
				}
				m.initTransactions()
			case paysetState:
				m.state = txnState
			}

		// navigate out of explorer views
		case key.Matches(msg, constants.Keys.Back):
			switch m.state {
			case paysetState:
				m.state = blockState
				m.initBlocks()
			case txnState:
				m.state = paysetState
			}
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case BlocksMsg:
		m.blocks = msg.blocks
		cmds = append(cmds, tea.Batch(
			tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
				// TODO: skip during catchup? Or make more/less frequent?
				return m.getLatestBlockHeaders()
			})))
	}

	t, tableCmd := m.table.Update(msg)
	m.table = t
	cmds = append(cmds, tableCmd)

	switch m.state {
	case blockState:
		m, updateCmd = m.UpdateBlocks(msg)
		return m, tea.Batch(append(cmds, updateCmd)...)
	case paysetState:
		return m, nil
	case txnState:
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case blockState, paysetState:
		return m.style.Bottom.Render(m.table.View())
	case txnState:
		return ""
	}
	return ""
}
