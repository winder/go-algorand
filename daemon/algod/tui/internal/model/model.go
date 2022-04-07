package model

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/about"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/accounts"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/configs"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/explorer"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/footer"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/status"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/tabs"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
)

const (
	initialWidth  = 80
	initialHeight = 50
)

type activeComponent int

const (
	explorerTab activeComponent = iota
	utilitiesTab
	accountTab
	configTab
	helpTab
)

type Model struct {
	Status        status.Model
	Accounts      accounts.Model
	Tabs          tabs.Model
	BlockExplorer explorer.Model
	Configs       configs.Model
	Utilities     tea.Model
	About         tea.Model
	Help          help.Model
	Footer        footer.Model

	network algod.NetworkMsg

	styles *style.Styles

	active activeComponent
	Server *algod.Server
	// remember the last resize so we can re-send it when selecting a different bottom component.
	lastResize tea.WindowSizeMsg
}

func New(s *algod.Server) Model {
	styles := style.DefaultStyles()
	tab := tabs.New([]string{"EXPLORER", "UTILITIES", "ACCOUNTS", "CONFIGURATION", "HELP"})
	// The tab content is the only flexible element.
	// This means the height must grow or shrink to fill the available
	// window height. It has access to the absolute height but needs to
	// be informed about the space used by other elements.
	tabContentMargin := style.TopHeight + tab.Height() + 2 /* +2 for footer/help */
	return Model{
		active:        explorerTab,
		Server:        s,
		styles:        styles,
		Status:        status.New(s, styles),
		Tabs:          tab,
		BlockExplorer: explorer.NewModel(s, styles, initialWidth, 0, initialHeight, tabContentMargin),
		Configs:       configs.New(s, tabContentMargin),
		Accounts:      accounts.NewModel(s, styles, initialHeight, tabContentMargin),
		Help:          help.New(),
		Footer:        footer.New(styles),
		About:         about.New(tabContentMargin, about.GetHelpContent()),
		Utilities:     about.New(tabContentMargin, about.GetUtilsContent()),
	}
}
