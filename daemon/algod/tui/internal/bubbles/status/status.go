package status

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
	"github.com/algorand/go-algorand/node"
)

type Model struct {
	Status  node.StatusReport
	Network algod.NetworkMsg
	Err     error

	style             *style.Styles
	server            *algod.Server
	progress          progress.Model
	processedAcctsPct float64
	verifiedAcctsPct  float64
	acquiredBlksPct   float64
}

func New(server *algod.Server, style *style.Styles) Model {
	return Model{
		style:    style,
		server:   server,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		algod.GetNetworkCmd(m.server),
		algod.GetStatusCmd(m.server),
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case algod.StatusMsg:
		if msg.Error != nil {
			m.Err = fmt.Errorf("error fetching status: %w", msg.Error)
			return m, tea.Quit
		}
		m.Status = msg.Status
		if m.Status.CatchpointCatchupTotalAccounts > 0 {
			m.processedAcctsPct = float64(m.Status.CatchpointCatchupProcessedAccounts) / float64(m.Status.CatchpointCatchupTotalAccounts)
			m.verifiedAcctsPct = float64(m.Status.CatchpointCatchupVerifiedAccounts) / float64(m.Status.CatchpointCatchupTotalAccounts)
		}
		if m.Status.CatchpointCatchupTotalBlocks > 0 {
			m.processedAcctsPct = 1
			m.verifiedAcctsPct = 1
			m.acquiredBlksPct = float64(m.Status.CatchpointCatchupAcquiredBlocks) / float64(m.Status.CatchpointCatchupTotalBlocks)
		}
		return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
			return algod.GetStatusCmd(m.server)()
		})

	case algod.NetworkMsg:
		m.Network = msg

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func formatVersion(v string) string {
	i := strings.LastIndex(v, "/")
	if i != 0 {
		i++
	}
	return v[i:]
}

func formatNextVersion(last, next string, round uint64) string {
	if last == next {
		return "N/A"
	}
	return strconv.FormatUint(round, 10)
}

func writeProgress(b *strings.Builder, prefix string, progress progress.Model, pct float64) {
	b.WriteString(prefix)
	b.WriteString(progress.ViewAs(pct))
	b.WriteString("\n")
}

func (m Model) View() string {
	bold := m.style.StatusBoldText
	key := m.style.BottomListItemKey.Copy().MarginLeft(0)
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("%s %s\n", bold.Render("Network:"), m.Network.GenesisID))
	builder.WriteString(fmt.Sprintf("%s %s\n", bold.Render("Genesis:"), m.Network.GenesisHash))
	// status
	if (m.Status != node.StatusReport{}) {
		nextVersion := formatNextVersion(
			string(m.Status.LastVersion),
			string(m.Status.NextVersion),
			uint64(m.Status.NextVersionRound))
		report := strings.Builder{}

		switch {
		case m.Status.Catchpoint != "":
			// Catchpoint view
			report.WriteString(fmt.Sprintf("Catchpoint: %s\n\n", m.Status.Catchpoint))
			var catchupStatus string
			switch {
			case m.Status.CatchpointCatchupAcquiredBlocks > 0:
				catchupStatus = fmt.Sprintf("Verifying accounts: %d / %d\n", m.Status.CatchpointCatchupAcquiredBlocks, m.Status.CatchpointCatchupTotalBlocks)
			case m.Status.CatchpointCatchupVerifiedAccounts > 0:
				catchupStatus = fmt.Sprintf("Verifying accounts: %d / %d\n", m.Status.CatchpointCatchupVerifiedAccounts, m.Status.CatchpointCatchupTotalAccounts)
			case m.Status.CatchpointCatchupProcessedAccounts > 0:
				catchupStatus = fmt.Sprintf("Downloading accounts: %d / %d\n", m.Status.CatchpointCatchupProcessedAccounts, m.Status.CatchpointCatchupTotalAccounts)
			}
			report.WriteString(bold.Render(catchupStatus))
			report.WriteString("\n")
			writeProgress(&report, "Downloading accounts: ", m.progress, m.processedAcctsPct)
			writeProgress(&report, "Processing accounts:  ", m.progress, m.verifiedAcctsPct)
			writeProgress(&report, "Downloading blocks:   ", m.progress, m.acquiredBlksPct)
		default:
			report.WriteString(fmt.Sprintf("Current round:   %s\n", key.Render(strconv.FormatUint(uint64(m.Status.LastRound), 10))))
			report.WriteString(fmt.Sprintf("Block wait time: %s\n", m.Status.TimeSinceLastRound()))
			report.WriteString(fmt.Sprintf("Sync time:       %s\n", m.Status.SynchronizingTime))
			// TODO: Display consensus upgrade progress
			if m.Status.LastVersion == m.Status.NextVersion {
				// no upgrade in progress
				report.WriteString(fmt.Sprintf("Protocol:        %s\n", formatVersion(string(m.Status.LastVersion))))
				report.WriteString(fmt.Sprintf("                 %s\n", bold.Render("No upgrade in progress.")))
			} else {
				// upgrade in progress
				report.WriteString(fmt.Sprintf("%s\n", bold.Render("Consensus Upgrade Pending")))
				report.WriteString(fmt.Sprintf("Current Protocol: %s\n", formatVersion(string(m.Status.LastVersion))))
				report.WriteString(fmt.Sprintf("Next Protocol:    %s\n", formatVersion(string(m.Status.NextVersion))))
				report.WriteString(fmt.Sprintf("Upgrade round:    %s\n", nextVersion))

			}
		}

		builder.WriteString(report.String())
	}

	return m.style.Status.Render(builder.String())
}
