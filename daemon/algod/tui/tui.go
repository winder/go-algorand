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

// Package tui contains a terminal UI started within the context of algod.
// Other components may need to be added to other packages to gain access to
// private data.
package tui

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/algorand/go-algorand/daemon/algod"
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
)

const host = "localhost"

func teaHandler(_ ssh.Session) (tea.Model, []tea.ProgramOption) {
	return model.New(algodServer), []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
}

var algodServer *algod.Server

// Start ...
func Start(s *algod.Server, port uint64) {
	if port == 0 {
		// Run directly
		p := tea.NewProgram(model.New(s), tea.WithAltScreen(), tea.WithMouseCellMotion())
		if err := p.Start(); err != nil {
			fmt.Printf("Error in UI: %v", err)
			os.Exit(1)
		}

		fmt.Printf("\nUI Terminated, shutting down node.\n")
		os.Exit(0)
	}

	// Run on ssh server.
	algodServer = s
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	sshServer, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(path.Join(dirname, ".ssh/term_info_ed25519")),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = sshServer.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := sshServer.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
