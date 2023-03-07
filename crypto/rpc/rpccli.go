package rpc

import (
	"fmt"
	"net/rpc"

	"github.com/bserdar/took/crypto"
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
	var rsp crypto.NameResponse
	cli.Call("RequestProcessor.GetName", &crypto.NameRequest{}, &rsp)
	if rsp.Name != name {
		return nil, fmt.Errorf("Server initialized for another file")
	}
	return &RequestProcessorClient{cli: cli}, nil
}

// Init sets up the server with the given password
func (s *RequestProcessorClient) Init(password, name string) (string, error) {
	var rsp crypto.InitResponse
	err := s.cli.Call("RequestProcessor.Init", &crypto.InitRequest{Password: password, Name: name}, &rsp)
	return rsp.AuthKey, err
}

// Login validates the password with the authKey
func (s *RequestProcessorClient) Login(password, authKey, name string) error {
	var rsp crypto.LoginResponse
	err := s.cli.Call("RequestProcessor.Login", &crypto.LoginRequest{Password: password,
		AuthKey: authKey,
		Name:    name}, &rsp)
	return err
}

// Encrypt a block of data
func (s *RequestProcessorClient) Encrypt(data string) (string, error) {
	var rsp crypto.DataResponse
	err := s.cli.Call("RequestProcessor.Encrypt", &crypto.DataRequest{Data: data}, &rsp)
	return rsp.Data, err
}

// Decrypt a block of data
func (s *RequestProcessorClient) Decrypt(data string) (string, error) {
	var rsp crypto.DataResponse
	err := s.cli.Call("RequestProcessor.Decrypt", &crypto.DataRequest{Data: data}, &rsp)
	return rsp.Data, err
}

// Ping the server
func (s *RequestProcessorClient) Ping() error {
	var rsp crypto.PingResponse
	return s.cli.Call("RequestProcessor.Ping", &crypto.PingRequest{}, &rsp)
}
