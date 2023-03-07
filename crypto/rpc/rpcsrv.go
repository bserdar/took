package rpc

import (
	"net"
	"net/rpc"
	"time"

	"github.com/bserdar/took/crypto"
)

// Server serves the encryption service via the unix domain socket
func Server(socketName, password, authKey, name string, idleTimeout time.Duration) error {
	server, err := crypto.NewServer(password, authKey)
	if err != nil {
		return err
	}
	tmr := time.NewTimer(idleTimeout)
	processor := crypto.NewRequestProcessor(server, func() {
		if !tmr.Stop() {
			<-tmr.C
		}
		tmr.Reset(idleTimeout)
	}, name)
	rpc.Register(&processor)
	listener, err := net.Listen("unix", socketName)
	if err != nil {
		return err
	}
	// Close the listener after a timeout
	go func() {
		<-tmr.C
		listener.Close()
	}()
	rpc.Accept(listener)
	return nil
}
