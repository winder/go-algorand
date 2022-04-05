package explorer

import (
	"fmt"
	"io"
	"strconv"
	"time"

	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/algorand/go-algorand/data/bookkeeping"
)

// blockItem is used by the list bubble.
type blockItem struct {
	*bookkeeping.Block
}

// Hacked these in to workaround missing style options in table model
var inactive = lipgloss.NewStyle()
var active = inactive.Copy().Foreground(lipgloss.Color("#B083EA")).Bold(true)
var key = inactive.Copy().Width(10).Foreground(lipgloss.Color("#A3A322")).Bold(true)

func (i blockItem) Render(w io.Writer, model table.Model, index int) {
	var cursor string
	if index == model.Cursor() {
		cursor = "> "
	} else {
		cursor = "  "
	}

	cursor = active.Render(cursor)
	round := key.Render(strconv.FormatUint(uint64(i.Round()), 10))
	rest := fmt.Sprintf("\t%d\t%d\t%d\t%d",
		len(i.Payset),
		12,
		13,
		14)
	if index == model.Cursor() {
		rest = active.Render(rest)
	} else {
		rest = inactive.Render(rest)
	}
	fmt.Fprintf(w, "%s%s%s\n", cursor, round, rest)
}

type blocks []blockItem

func (m *Model) initBlock() {
	m.blockPerPage = 25
	//func newBlockModel(node *node.AlgorandFullNode, styles *style.Styles, width, widthMargin, height, heightMargin int) blockModel {
	t := table.New([]string{"  ROUND", "Txns", "Axfer", "Acfg", "Appl"}, 0, 0)
	t.KeyMap.Up.SetKeys(append(t.KeyMap.Up.Keys(), "k")...)
	t.KeyMap.Down.SetKeys(append(t.KeyMap.Down.Keys(), "j")...)
	m.table = t
	m.SetSize(m.width, m.height)
}

type BlocksMsg struct {
	blocks []blockItem
	err    error
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

func (m Model) InitBlock() tea.Cmd {
	return m.getLatestBlockHeaders
}

func (m *Model) SetSize(width, height int) {
	m.table.SetSize(width-m.widthMargin, height-m.heightMargin)
}

func (m Model) UpdateBlock(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BlocksMsg:
		m.blocks = msg.blocks
		var rows []table.Row
		for _, b := range m.blocks {
			rows = append(rows, b)
		}

		m.table.SetRows(rows)
		return m, tea.Batch(
			tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
				// TODO: skip during catchup? Or make more/less frequent?
				return m.getLatestBlockHeaders()
			}),
		)
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}

	t, tableCmd := m.table.Update(msg)
	m.table = t

	return m, tea.Batch(tableCmd)
}

func (m Model) ViewBlock() string {
	return m.style.Bottom.Render(m.table.View())
}
