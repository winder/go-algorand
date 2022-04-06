package explorer

import (
	"fmt"
	"io"
	"strconv"

	table "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/protocol"
)

// blockItem is used by the list bubble.
type blockItem struct {
	*bookkeeping.Block
}

// Hacked these in to workaround missing style options in table model
var inactiveStyle = lipgloss.NewStyle()
var activeStyle = inactiveStyle.Copy().Foreground(lipgloss.Color("#B083EA")).Bold(true)
var keyStyle = inactiveStyle.Copy().Width(10).Foreground(lipgloss.Color("#A3A322")).Bold(true)

var blockTableHeader = []string{"  ROUND", "Txns", "Pay", "[algos transferred]", "Axfer", "Acfg", "Afrz", "[Unique assets]", "Appl", "[Unique apps]"}

func computeBlockRow(b blockItem) string {
	types := make(map[protocol.TxType]uint)
	var paymentsTotal uint64
	assets := make(map[uint64]struct{})
	apps := make(map[uint64]struct{})
	for _, tx := range b.Payset {
		types[tx.Txn.Type]++

		switch tx.Txn.Type {
		case protocol.PaymentTx:
			paymentsTotal += tx.Txn.PaymentTxnFields.Amount.Raw
		case protocol.ApplicationCallTx:
			id := uint64(tx.Txn.ApplicationCallTxnFields.ApplicationID)
			if id == 0 {
				id = uint64(tx.ApplyData.ApplicationID)
			}
			if id == 0 {
				break
			}
			if _, ok := apps[id]; !ok {
				apps[id] = struct{}{}
			}
		case protocol.AssetTransferTx:
			fallthrough
		case protocol.AssetFreezeTx:
			fallthrough
		case protocol.AssetConfigTx:
			id := uint64(tx.Txn.AssetTransferTxnFields.XferAsset)
			if id == 0 {
				id = uint64(tx.ApplyData.ConfigAsset)
			}
			if id == 0 {
				id = uint64(tx.Txn.AssetConfigTxnFields.ConfigAsset)
			}
			if id == 0 {
				id = uint64(tx.Txn.AssetFreezeTxnFields.FreezeAsset)
			}
			if id == 0 {
				break
			}
			if _, ok := assets[id]; !ok {
				assets[id] = struct{}{}
			}
		}
	}

	return fmt.Sprintf("\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d",
		len(b.Payset),
		types[protocol.PaymentTx],
		paymentsTotal,
		types[protocol.AssetTransferTx],
		types[protocol.AssetConfigTx],
		types[protocol.AssetFreezeTx],
		len(assets),
		types[protocol.ApplicationCallTx],
		len(apps))

}

func (i blockItem) Render(w io.Writer, model table.Model, index int) {
	var cursor string
	if index == model.Cursor() {
		cursor = "> "
	} else {
		cursor = "  "
	}

	cursor = activeStyle.Render(cursor)
	round := keyStyle.Render(strconv.FormatUint(uint64(i.Round()), 10))
	rest := computeBlockRow(i)
	if index == model.Cursor() {
		rest = activeStyle.Render(rest)
	} else {
		rest = inactiveStyle.Render(rest)
	}
	fmt.Fprintf(w, "%s%s%s\n", cursor, round, rest)
}

func (m *Model) updateBlockTable() {
	if len(m.blocks) <= 0 {
		return
	}

	var rows []table.Row
	for _, b := range m.blocks {
		rows = append(rows, b)
	}

	m.table.SetRows(rows)
}

func (m *Model) initBlocks() {
	m.blockPerPage = 25
	t := table.New(blockTableHeader, 0, 0)
	t.KeyMap.Up.SetKeys(append(t.KeyMap.Up.Keys(), "k")...)
	t.KeyMap.Down.SetKeys(append(t.KeyMap.Down.Keys(), "j")...)
	t.Styles.Title = m.style.StatusBoldText
	m.table = t
	m.SetSize(m.width, m.height)
	m.updateBlockTable()
}

func (m Model) UpdateBlocks(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case BlocksMsg:
		if m.state == blockState {
			m.updateBlockTable()
		}
	}

	return m, nil
}
