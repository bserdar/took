package rpc

import (
	"fmt"
	"net/rpc"

	"github.com/bserdar/took/crypta"
)

// RequestProcessorClient wraps the connection to the rpc server
type RequestProcessorClient struct {
	cli *rpc.Client
}

// NewRequestProcessorClient creates a new client to the encrypt/decrypt agent
func NewRequestProcessorClient(network, address, name string) (*RequestProcessorClient, error) {
	cli, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	var rsp crypta.NameResponse
	cli.Call("RequestProcessor.GetName", &crypta.NameRequest{}, &rsp)
	if rsp.Name != name {
		return nil, fmt.Errorf("Server initialized for another file")
	}
	return &RequestProcessorClient{cli: cli}, nil
}

// Init sets up the server with the given password
func (s *RequestProcessorClient) Init(password, name string) (string, error) {
	var rsp crypta.InitResponse
	err := s.cli.Call("RequestProcessor.Init", &crypta.InitRequest{Password: password, Name: name}, &rsp)
	return rsp.AuthKey, err
}

// Login validates the password with the authKey
func (s *RequestProcessorClient) Login(password, authKey, name string) error {
	var rsp crypta.LoginResponse
	err := s.cli.Call("RequestProcessor.Login", &crypta.LoginRequest{Password: password,
		AuthKey: authKey,
		Name:    name}, &rsp)
	return err
}

// Encrypt a block of data
func (s *RequestProcessorClient) Encrypt(data string) (string, error) {
	var rsp crypta.DataResponse
	err := s.cli.Call("RequestProcessor.Encrypt", &crypta.DataRequest{Data: data}, &rsp)
	return rsp.Data, err
}

// Decrypt a block of data
func (s *RequestProcessorClient) Decrypt(data string) (string, error) {
	var rsp crypta.DataResponse
	err := s.cli.Call("RequestProcessor.Decrypt", &crypta.DataRequest{Data: data}, &rsp)
	return rsp.Data, err
}

// Ping the server
func (s *RequestProcessorClient) Ping() error {
	var rsp crypta.PingResponse
	return s.cli.Call("RequestProcessor.Ping", &crypta.PingRequest{}, &rsp)
}
