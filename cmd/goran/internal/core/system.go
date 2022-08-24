package core

import (
	"os"
	"time"
)

type System struct {
}

type SystemShutdownArgs struct {
	SecondsDelay int
}

type SystemShutdownReply struct {
	Received bool
}

// Shutdown the RPC server.
func (s *System) Shutdown(args SystemShutdownArgs, reply *SystemShutdownReply) error {
	reply.Received = true
	go func() {
		<-time.After(time.Duration(args.SecondsDelay) * time.Second)
		os.Exit(0)
	}()
	return nil
}
