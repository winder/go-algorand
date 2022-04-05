package constants

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit         key.Binding
	Catchup      key.Binding
	AbortCatchup key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Catchup, k.AbortCatchup}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

var Keys = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit")),
	Catchup: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "start fast catchup")),
	AbortCatchup: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "abort catchup")),
}
