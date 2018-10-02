package crypta

import (
	"bytes"
	"testing"
)

func TestInit(t *testing.T) {
	server, err := InitServer("password")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	authKey, err := server.GetAuthKey()
	if err != nil {
		t.Errorf("Cannot get auth key: %s", err.Error())
		return
	}
	t.Logf("authKey: %s", authKey)

	server2, err := NewServer("password", authKey)
	if err != nil {
		t.Errorf("Cannot reauth: %s", err.Error())
	}

	if bytes.Compare(server.key, server2.key) != 0 {
		t.Errorf("Keys are different")
	}
	if bytes.Compare(server.salt, server2.salt) != 0 {
		t.Errorf("salts are different")
	}
	if bytes.Compare(server.iv, server2.iv) != 0 {
		t.Errorf("ivs are different")
	}

	_, err = NewServer("blah", authKey)
	if err == nil {
		t.Errorf("Invalid password should've failed")
	}
}

func TestEncDec(t *testing.T) {
	server, err := InitServer("password")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	authKey, err := server.GetAuthKey()
	if err != nil {
		t.Errorf("Cannot get auth key: %s", err.Error())
		return
	}
	data := make([]byte, 512)
	rndb(data)
	out, err := server.Encrypt(data)
	if err != nil {
		t.Errorf("Unexpected: %s", err.Error())
	}
	check, err := server.Decrypt(out)
	if err != nil {
		t.Errorf("Unexpected: %s", err.Error())
	}
	if bytes.Compare(data, check) != 0 {
		t.Errorf("in!=out in=%s out=%s", data, check)
	}

	server2, err := NewServer("password", authKey)
	check, err = server2.Decrypt(out)
	if err != nil {
		t.Errorf("Unexpected: %s", err.Error())
	}
	if bytes.Compare(data, check) != 0 {
		t.Errorf("in!=out in=%s out=%s", data, check)
	}

}
