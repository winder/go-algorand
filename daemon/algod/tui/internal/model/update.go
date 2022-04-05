package model

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod/tui/internal/constants"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keys.Quit):
			return m, tea.Quit
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

	return m, tea.Batch(statusCommand, accountsCommand, explorerCommand)
}
