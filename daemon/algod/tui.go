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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/node"
)

var (
	AddressList = []string{
		//"ZONEGRWBV3Q7JA6RHAN4EMAX6ICIVZX2C6U65DCNHYLIL4PUBB7O6DOSBI",
		//"XFYAYSEGQIY2J3DCGGXCPXY5FGHSVKM3V4WCNYCLKDLHB7RYDBU233QB5M",
		//"WNEJFT6HTAX3CQ6YOPIY65AKYCBQM6BLV4S5OP54VH76OP33LOL2MYGSIM",
		//"GULDQIEZ2CUPBSHKXRWUW7X3LCYL44AI5GGSHHOQDGKJAZ2OANZJ43S72U",
		"57QZ4S7YHTWPRAM3DQ2MLNSVLAQB7DTK4D7SUNRIEFMRGOU7DMYFGF55BY",
		//"ETGSQKACKC56JWGMDAEP5S2JVQWRKTQUVKCZTMPNUGZLDVCWPY63LSI3H4",
		"J4AEINCSSLDA7LNBNWM4ZXFCTLTOZT5LG3F5BLMFPJYGFWVCMU37EZI2AM",
	}
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

type AccountStatusMsg map[basics.Address]uint64

func GetAccountStatusMsg(s *Server) tea.Cmd {
	return func() tea.Msg {
		currentNode := GetNode(s)
		ledger := currentNode.Ledger()

		rval := make(AccountStatusMsg)

		for _, a := range AddressList {
			currentAddress, err := basics.UnmarshalChecksumAddress(a)

			if err != nil {
				continue
			}

			record, _, _, err := ledger.LookupLatest(currentAddress)

			rval[currentAddress] = record.MicroAlgos.Raw
		}

		return rval
	}
}

func StartFastCatchup(s *Server) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/testnet/latest.catchpoint")
		if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		catchpoint := strings.Replace(string(body), "#", "%23", 1)

		//start fast catchup
		url := fmt.Sprintf("http://localhost:8080/v2/catchup/%s", catchpoint)
		url = url[:len(url)-1] // remove \n
		apiToken, err := os.ReadFile(path.Join(os.Getenv("ALGORAND_DATA"), "algod.admin.token"))
		if err != nil {
			panic(err)
		}
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("X-Algo-Api-Token", string(apiToken))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		s, err := s.node.Status()
		return StatusMsg{
			Status: s,
			Error:  err,
		}
	}
}

func StopFastCatchup(s *Server) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/testnet/latest.catchpoint")
		if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		catchpoint := strings.Replace(string(body), "#", "%23", 1)

		//start fast catchup
		url := fmt.Sprintf("http://localhost:8080/v2/catchup/%s", catchpoint)
		url = url[:len(url)-1] // remove \n
		apiToken, err := os.ReadFile(path.Join(os.Getenv("ALGORAND_DATA"), "algod.admin.token"))
		if err != nil {
			panic(err)
		}
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("X-Algo-Api-Token", string(apiToken))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

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
