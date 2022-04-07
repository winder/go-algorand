package about

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type Model struct {
	width        int
	height       int
	heightMargin int
	viewport     viewport.Model
}

func getHelpContent() string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(45),
		glamour.WithEmoji(),
	)

	builder := strings.Builder{}
	title := `# Node UI :smiley_cat:`
	desc := `An awesome node TUI.`
	shortcuts := ` ## Shortcut Keys 
- **F**ast catchup
- **A**bort fast catchup
- **S**end payment
- **D**elete block
- **H**ack relay
- ...`

	status := `## Node Status
| Key  | Description |
| ------ | ----- |
| Network  | this   |
| Genesis | is   |
| Current round  | a   |
| Sync time | demo   |
| ... | !   |

Start fast catchup to see catchup status.
`

	explorer := `## Block Explorer
*real time blocks and transaction details*
`
	accounts := `## Accounts
This section display local account details. 
`

	configs := `## Node Configurations
**Your node settings** `

	code := `{
"Version": 16,
"AccountsRebuildSynchronousMode": 1,
"AnnounceParticipationKey": true,
"Archival": false,
"BaseLoggerDebugLevel": 4,
"BroadcastConnectionsLimit": -1,
"CadaverSizeTarget": 1073741824,
...
}
`
	builder.WriteString(fmt.Sprintf("%s\n\n", title))
	builder.WriteString(fmt.Sprintf("%s\n\n", desc))
	builder.WriteString(fmt.Sprintf("%s\n\n", shortcuts))
	builder.WriteString(fmt.Sprintf("%s\n\n", status))
	builder.WriteString(fmt.Sprintf("%s\n\n", explorer))
	builder.WriteString(fmt.Sprintf("%s\n\n", accounts))
	builder.WriteString(fmt.Sprintf("%s\n```json\n%s\n```\n\n", configs, code))

	content, _ := r.Render(builder.String())

	return content
}

func New(heightMargin int) Model {
	m := Model{
		heightMargin: heightMargin,
		viewport:     viewport.New(0, 0),
	}
	m.setSize(80, 20)
	m.viewport.SetContent(getHelpContent())
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
	}
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("%s", m.viewport.View()))
	return builder.String()
}

func (m *Model) setSize(width, height int) {
	m.viewport.Width = width
	m.viewport.Height = height - m.heightMargin
}
