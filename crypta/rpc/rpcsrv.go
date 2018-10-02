package rpc

import (
	"net"
	"net/rpc"
	"time"

	"github.com/bserdar/took/crypta"
)

// RPCServer serves the encryption service via the unix domain socket
func RPCServer(socketName, password, authKey string, idleTimeout time.Duration) error {
	server, err := crypta.NewServer(password, authKey)
	if err != nil {
		return err
	}
	tmr := time.NewTimer(idleTimeout)
	processor := crypta.NewRequestProcessor(server, func() {
		if !tmr.Stop() {
			<-tmr.C
		}
		tmr.Reset(idleTimeout)
	})
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
