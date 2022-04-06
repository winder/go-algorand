package model

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/accounts"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/explorer"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/footer"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/bubbles/status"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
)

const (
	// MaxTopBoxHeight is the height of the top boxes. Hard coded to avoid dynamic margins.
	MaxTopBoxHeight = style.TopHeight
	initialWidth    = 80
	initialHeight   = 50
)

type Model struct {
	Server        *algod.Server
	Status        status.Model
	Accounts      accounts.Model
	BlockExplorer explorer.Model
	Help          help.Model
	Footer        footer.Model

	network algod.NetworkMsg

	styles *style.Styles

	// remember the last resize so we can re-send it when selecting a different bottom component.
	lastResize tea.WindowSizeMsg
}

func New(s *algod.Server) Model {
	styles := style.DefaultStyles()
	return Model{
		Server:        s,
		styles:        styles,
		Status:        status.New(s, styles),
		BlockExplorer: explorer.NewModel(s, styles, initialWidth, 0, initialHeight, MaxTopBoxHeight /* Max(status.height, account.height) */),
		Accounts:      accounts.NewModel(s),
		Help:          help.New(),
		Footer:        footer.New(styles),
	}
}
