package model

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/constants"
)

func networkFromID(genesisID string) string {
	return strings.Split(genesisID, "-")[0]
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case algod.NetworkMsg:
		m.network = msg
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.Keys.Catchup):
			return m, algod.StartFastCatchup(m.Server, networkFromID(m.Status.Network.GenesisID))
		case key.Matches(msg, constants.Keys.AbortCatchup):
			return m, algod.StopFastCatchup(m.Server, networkFromID(m.Status.Network.GenesisID))
		}

	case tea.WindowSizeMsg:
		m.lastResize = msg
	}

	var statusCommand tea.Cmd
	m.Status, statusCommand = m.Status.Update(msg)

	var accountsCommand tea.Cmd
	m.Accounts, accountsCommand = m.Accounts.Update(msg)

	var explorerCommand tea.Cmd
	m.BlockExplorer, explorerCommand = m.BlockExplorer.Update(msg)

	var footerCommand tea.Cmd
	m.Footer, footerCommand = m.Footer.Update(msg)

	return m, tea.Batch(statusCommand, accountsCommand, explorerCommand, footerCommand)
}
