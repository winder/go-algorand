package accounts

import (
	"fmt"
	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/style"
	"github.com/algorand/go-algorand/data/basics"
	tea "github.com/charmbracelet/bubbletea"
	"sort"
	"strings"
	"time"
)

type balance struct {
	MicroAlgos uint64
	TimeStamp  time.Time
}

type account struct {
	CurrentBalance balance
	BalanceHistory []balance
}

type Model struct {
	Accounts map[basics.Address]*account

	server *algod.Server
	Err    error
	style             *style.Styles
}

func NewModel(server *algod.Server, style *style.Styles) Model {
	rval := Model{
		Accounts: make(map[basics.Address]*account),
		server:   server,
		style: style,
	}

	for _, a := range algod.AddressList {

		currentAddress, err := basics.UnmarshalChecksumAddress(a)

		if err != nil {
			continue
		}

		rval.Accounts[currentAddress] = &account{BalanceHistory: []balance{
			{0, time.Now()},
			{0, time.Now()},
			{0, time.Now()},
		}}

	}

	return rval
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		algod.GetAccountStatusMsg(m.server))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case algod.AccountStatusMsg:

		for msgAddress, msgBalance := range msg {

			// Don't update if the balance didn't change
			if msgBalance == m.Accounts[msgAddress].CurrentBalance.MicroAlgos {
				return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
					return algod.GetAccountStatusMsg(m.server)()
				})
			}

			tmpList := m.Accounts[msgAddress].BalanceHistory

			// Prepend the balance
			tmpList = append([]balance{m.Accounts[msgAddress].CurrentBalance}, tmpList...)
			if len(tmpList) > 3 {
				tmpList = tmpList[:3]
			}

			m.Accounts[msgAddress].BalanceHistory = tmpList

			m.Accounts[msgAddress].CurrentBalance = balance{
				MicroAlgos: msgBalance,
				TimeStamp:  time.Now(),
			}

			return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return algod.GetAccountStatusMsg(m.server)()
			})
		}

	}

	return m, nil
}

func (m Model) View() string {
	builder := strings.Builder{}

	keys := make([]string, 0, len(m.Accounts))
	for k := range m.Accounts {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)

	for _, account := range keys {
		accountType, _ := basics.UnmarshalChecksumAddress(account)
		v := m.Accounts[accountType]
		builder.WriteString(fmt.Sprintf("Address: %s\n", account))
		builder.WriteString(fmt.Sprintf("  Balance: %f Algos @ %s\n", float64(v.CurrentBalance.MicroAlgos)/1000000.0, v.CurrentBalance.TimeStamp.Format("2006-01-02 15:04:05.1234")))
		for _, a := range v.BalanceHistory {
			if a.MicroAlgos == 0 {
				builder.WriteString(fmt.Sprintf("     --> -- Algos @ --\n"))
			} else {
				builder.WriteString(fmt.Sprintf("     --> %f Algos @ %s\n", float64(a.MicroAlgos)/1000000, a.TimeStamp.Format("2006-01-02 15:04:05.1234")))
			}
		}

		builder.WriteString("-----------------------\n")
	}

	return m.style.Account.Render(builder.String())
}
