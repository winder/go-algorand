// DEPRECATED, USING TABLE INSTEAD OF LIST.
// KEEPING THIS FOR REFERENCE IN CASE A LIST IS NEEDED ELSEWHERE.
package explorer

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"

	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/node"
)

// blockItem is used by the list bubble.
type blockItemList struct {
	*bookkeeping.Block
}

func (i blockItemList) Title() string {
	return fmt.Sprintf("Txs: %-5d Asset: %-5d App: %-5d", len(i.Payset), 99, 101)
}

func (i blockItemList) FilterValue() string { return i.Title() }

// itemDelegate is used for rendering the list item.
type itemDelegate struct {
	style *style.Styles
}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(blockItemList)
	if !ok {
		return
	}

	leftMargin := d.style.BottomListItemSelector.GetMarginLeft() +
		d.style.BottomListItemSelector.GetWidth() +
		d.style.BottomListItemKey.GetMarginLeft() +
		d.style.BottomListItemKey.GetWidth() +
		d.style.BottomListItemInactive.GetMarginLeft()
	title := truncate.StringWithTail(i.Title(), uint(m.Width()-leftMargin), "…")
	id := strconv.FormatUint(uint64(i.Round()), 10)

	if index == m.Index() {
		fmt.Fprint(w, d.style.BottomListItemSelector.Render(">")+
			d.style.BottomListItemKey.Bold(true).Render(id)+
			d.style.BottomListItemActive.Render(title))
	} else {
		fmt.Fprint(w, d.style.BottomListItemSelector.Render(" ")+
			d.style.BottomListItemKey.Bold(true).Render(id)+
			d.style.BottomListItemInactive.Render(title))
	}
}

type blocksList []blockItemList

type blockModelList struct {
	width        int
	widthMargin  int
	height       int
	heightMargin int
	style        *style.Styles

	blockPerPage uint

	node *node.AlgorandFullNode

	blocks blocksList

	list list.Model
}

func newBlockModelList(node *node.AlgorandFullNode, styles *style.Styles, width, widthMargin, height, heightMargin int) blockModelList {
	l := list.New([]list.Item{}, itemDelegate{styles}, 0, 0)
	l.Title = "Block Explorer"
	l.Styles.Title = styles.BottomListTitle
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowTitle(true)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.Select(0)

	b := blockModelList{
		blockPerPage: 25,
		style:        styles,
		node:         node,
		list:         l,
		widthMargin:  widthMargin,
		heightMargin: heightMargin,
	}
	b.SetSize(width, height)
	return b
}

type BlocksMsgList struct {
	blocks []blockItemList
	err    error
}

func (b *blockModelList) getLatestBlockHeaders() tea.Msg {
	// TODO: Only fetch if needed, check current latest vs actual latest
	var result BlocksMsgList

	ledger := b.node.Ledger()
	latest := ledger.Latest()
	for b.blockPerPage > uint(len(result.blocks)) && latest > 0 {
		block, err := ledger.Block(latest)
		if err != nil {
			result.err = err
			return result
		}
		latest -= 1

		result.blocks = append(result.blocks, blockItemList{&block})
	}
	return result
}

func (b blockModelList) Init() tea.Cmd {
	return b.getLatestBlockHeaders
}

func (b *blockModelList) SetSize(width, height int) {
	b.width = width
	b.height = height
	b.list.SetSize(width-b.widthMargin, height-b.heightMargin)
	b.list.Styles.PaginationStyle = b.style.BottomPaginator.Copy().Width(width - b.widthMargin)
}

func (b *blockModelList) Update(msg tea.Msg) (*blockModelList, tea.Cmd) {
	switch msg := msg.(type) {
	case BlocksMsgList:
		b.blocks = msg.blocks
		var items []list.Item
		//var titems []table.Row
		for _, b := range b.blocks {
			items = append(items, b)
			//titems = append(titems, b.toRow())
		}

		return b, tea.Batch(
			b.list.SetItems(items),
			tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
				// TODO: skip during catchup? Or make more frequent?
				return b.getLatestBlockHeaders()
			}),
		)
	case tea.WindowSizeMsg:
		b.SetSize(msg.Width, msg.Height)
	}

	l, listCmd := b.list.Update(msg)
	b.list = l

	return b, tea.Batch(listCmd)
}

func (b blockModelList) View() string {
	return b.style.Bottom.Render(b.list.View())
}
