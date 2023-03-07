package crypto

import (
	"encoding/base64"
	"errors"
)

// RequestProcessor is created after a successful login. It implements
// a simple RPC protocol betwee the agent and a client using the agent
// to encrypt/decrypt data
type RequestProcessor struct {
	srv  Server
	Name string
	ping func()
}

// InitRequest initializes the decryption server with the given password
type InitRequest struct {
	Name     string `json:"name"`
	Password string `json:"pwd"`
}

// InitResponse returns the AuthKey. This is a string to validate the decryption password
type InitResponse struct {
	AuthKey string `json:"auth"`
}

// LoginRequest sends the password and the authkey to the serevr to validate
type LoginRequest struct {
	Password string `json:"pwd"`
	AuthKey  string `json:"auth"`
	Name     string `json:"name"`
}

// LoginResponse returns the result of login password validation
type LoginResponse struct {
	Ok bool `json:"ok"`
}

// DataRequest contains data to be encrypted/decrypted
type DataRequest struct {
	Data string `json:"data"`
}

// DataResponse contains the encrypted/decrypted data
type DataResponse struct {
	Data string `json:"data"`
}

// NameRequest requests the name of the server
type NameRequest struct {
}

// NameResponse contains the name of the server
type NameResponse struct {
	Name string `json:"name"`
}

// PingRequest is ti check connection status and extend idle timeout
type PingRequest struct{}

// PingResponse is empty
type PingResponse struct{}

// NewRequestProcessor return a new RequestProcessor object using the given server
func NewRequestProcessor(server Server, pingFunc func(), name string) RequestProcessor {
	if pingFunc == nil {
		pingFunc = func() {}
	}
	return RequestProcessor{srv: server, ping: pingFunc, Name: name}
}

// Init initializes the processor with a password
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
	s.Name = req.Name
	response.AuthKey = a
	return nil
}

// Login validates the password and sets up the encryption server
func (s *RequestProcessor) Login(req LoginRequest, response *LoginResponse) error {
	s.ping()
	srv, err := NewServer(req.Password, req.AuthKey)
	if err != nil {
		return err
	}
	s.srv = srv
	s.Name = req.Name
	response.Ok = true
	return nil
}

// Encrypt a block of data
func (s *RequestProcessor) Encrypt(req DataRequest, response *DataResponse) error {
	s.ping()
	data, err := s.srv.Encrypt([]byte(req.Data))
	if err != nil {
		return err
	}
	response.Data = base64.StdEncoding.EncodeToString(data)
	return nil
}

// Decrypt a block of data
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

// Ping calls the ping function
func (s *RequestProcessor) Ping(req PingRequest, response *PingResponse) error {
	s.ping()
	return nil
}

// GetName returns the name of the server
func (s *RequestProcessor) GetName(req NameRequest, response *NameResponse) error {
	response.Name = s.Name
	return nil
}
