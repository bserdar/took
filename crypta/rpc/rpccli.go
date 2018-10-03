package rpc

import (
	"net/rpc"

	"github.com/bserdar/took/crypta"
)

// RequestProcessor is created after a successful login
type RequestProcessorClient struct {
	cli *rpc.Client
}

func NewRequestProcessorClient(network, address string) (*RequestProcessorClient, error) {
	cli, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &RequestProcessorClient{cli: cli}, nil
}

func (s *RequestProcessorClient) Init(password string) (string, error) {
	var rsp crypta.InitResponse
	err := s.cli.Call("RequestProcessor.Init", &crypta.InitRequest{Password: password}, &rsp)
	return rsp.AuthKey, err
}

func (s *RequestProcessorClient) Login(password, authKey string) error {
	var rsp crypta.LoginResponse
	err := s.cli.Call("RequestProcessor.Login", &crypta.LoginRequest{Password: password, AuthKey: authKey}, &rsp)
	return err
}

func (s *RequestProcessorClient) Encrypt(data string) (string, error) {
	var rsp crypta.DataResponse
	err := s.cli.Call("RequestProcessor.Encrypt", &crypta.DataRequest{Data: data}, &rsp)
	return rsp.Data, err
}

func (s *RequestProcessorClient) Decrypt(data string) (string, error) {
	var rsp crypta.DataResponse
	err := s.cli.Call("RequestProcessor.Decrypt", &crypta.DataRequest{Data: data}, &rsp)
	return rsp.Data, err
}

func (s *RequestProcessorClient) Ping() error {
	var rsp crypta.PingResponse
	return s.cli.Call("RequestProcessor.Ping", &crypta.PingRequest{}, &rsp)
}
