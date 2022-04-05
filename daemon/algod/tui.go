// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package algod

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/node"
)

type NetworkMsg struct {
	GenesisID   string
	GenesisHash crypto.Digest
}

func GetNetworkCmd(s *Server) tea.Cmd {
	return func() tea.Msg {
		return NetworkMsg{
			GenesisID:   s.node.GenesisID(),
			GenesisHash: s.node.GenesisHash(),
		}
	}
}

type StatusMsg struct {
	Status node.StatusReport
	Error  error
}

func GetStatusCmd(s *Server) tea.Cmd {
	return func() tea.Msg {
		s, err := s.node.Status()
		return StatusMsg{
			Status: s,
			Error:  err,
		}
	}
}

func GetNode(s *Server) *node.AlgorandFullNode {
	return s.node
}
