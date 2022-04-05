package explorer

import (
	"fmt"
	"io"
	"strconv"
	"time"

	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/node"
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

type blockModel struct {
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

func newBlockModel(node *node.AlgorandFullNode, styles *style.Styles, width, widthMargin, height, heightMargin int) blockModel {
	t := table.New([]string{"  ROUND", "Txns", "Axfer", "Acfg", "Appl"}, 0, 0)
	t.KeyMap.Up.SetKeys(append(t.KeyMap.Up.Keys(), "k")...)
	t.KeyMap.Down.SetKeys(append(t.KeyMap.Down.Keys(), "j")...)
	b := blockModel{
		blockPerPage: 25,
		style:        styles,
		node:         node,
		table:        t,
		widthMargin:  widthMargin,
		heightMargin: heightMargin,
	}
	b.SetSize(width, height)
	return b
}

type BlocksMsg struct {
	blocks []blockItem
	err    error
}

func (b *blockModel) getLatestBlockHeaders() tea.Msg {
	// TODO: Only fetch if needed, check current latest vs actual latest
	var result BlocksMsg

	ledger := b.node.Ledger()
	latest := ledger.Latest()
	for b.blockPerPage > uint(len(result.blocks)) && latest > 0 {
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

func (b blockModel) Init() tea.Cmd {
	return b.getLatestBlockHeaders
}

func (b *blockModel) SetSize(width, height int) {
	b.width = width
	b.height = height
	b.table.SetSize(width-b.widthMargin, height-b.heightMargin)
}

func (b *blockModel) Update(msg tea.Msg) (*blockModel, tea.Cmd) {
	switch msg := msg.(type) {
	case BlocksMsg:
		b.blocks = msg.blocks
		var rows []table.Row
		for _, b := range b.blocks {
			rows = append(rows, b)
		}

		b.table.SetRows(rows)
		return b, tea.Batch(
			tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
				// TODO: skip during catchup? Or make more/less frequent?
				return b.getLatestBlockHeaders()
			}),
		)
	case tea.WindowSizeMsg:
		b.SetSize(msg.Width, msg.Height)
	}

	t, tableCmd := b.table.Update(msg)
	b.table = t

	return b, tea.Batch(tableCmd)
}

func (b blockModel) View() string {
	return b.style.Bottom.Render(b.table.View())
}
