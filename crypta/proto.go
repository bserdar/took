package crypta

import (
	"encoding/base64"
	"errors"
)

// RequestProcessor is created after a successful login
type RequestProcessor struct {
	srv  Server
	ping func()
}

type InitRequest struct {
	Password string `json:"pwd"`
}

type InitResponse struct {
	AuthKey string `json:"auth"`
}

type LoginRequest struct {
	Password string `json:"pwd"`
	AuthKey  string `json:"auth"`
}

type LoginResponse struct {
	Ok bool `json:"ok"`
}

type DataRequest struct {
	Data string `json:"data"`
}

type DataResponse struct {
	Data string `json:"data"`
}

type PingRequest struct{}
type PingResponse struct{}

func NewRequestProcessor(server Server, pingFunc func()) RequestProcessor {
	if pingFunc == nil {
		pingFunc = func() {}
	}
	return RequestProcessor{srv: server, ping: pingFunc}
}

func (s *RequestProcessor) Init(req InitRequest, response *InitResponse) error {
	s.ping()
	if len(req.Password) == 0 {
		return errors.New("Empty password")
	}
	srv, err := InitServer(req.Password)
	if err != nil {
		return err
	}
	s.srv = srv
	a, err := srv.GetAuthKey()
	if err != nil {
		return err
	}
	response.AuthKey = a
	return nil
}

func (s *RequestProcessor) Login(req LoginRequest, response *LoginResponse) error {
	s.ping()
	srv, err := NewServer(req.Password, req.AuthKey)
	if err != nil {
		return err
	}
	s.srv = srv
	response.Ok = true
	return nil
}

func (s *RequestProcessor) Encrypt(req DataRequest, response *DataResponse) error {
	s.ping()
	data, err := s.srv.Encrypt([]byte(req.Data))
	if err != nil {
		return err
	}
	response.Data = base64.StdEncoding.EncodeToString(data)
	return nil
}

func (s *RequestProcessor) Decrypt(req DataRequest, response *DataResponse) error {
	s.ping()
	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return err
	}
	d, err := s.srv.Decrypt(data)
	if err != nil {
		return err
	}
	response.Data = string(d)
	return nil
}

func (s *RequestProcessor) Ping(req PingRequest, response *PingResponse) error {
	s.ping()
	return nil
}
