package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const saltSize = 32
const rndBlockSize = 32

// Encrypt/decrypt errors
var (
	ErrInvalidAuthKey  = errors.New("Invalid authKey")
	ErrInvalidPassword = errors.New("Invalid password")
)

// Server keeps all the information necessary to encrypt/decrypt data
type Server struct {
	key   []byte
	salt  []byte
	iv    []byte
	block cipher.Block
}

func rnd(n int) ([]byte, error) {
	ret := make([]byte, n)
	return ret, rndb(ret)
}

func rndb(out []byte) error {
	_, err := io.ReadFull(rand.Reader, out)
	return err
}

// InitServer creates a new encrypt/decrypt server using the given
// password. It returns the server containing a key and salt
func InitServer(password string) (Server, error) {
	ret := Server{}
	ret.salt = make([]byte, saltSize)
	err := rndb(ret.salt)
	if err != nil {
		return ret, err
	}
	ret.iv = make([]byte, aes.BlockSize)
	err = rndb(ret.iv)
	if err != nil {
		return ret, err
	}
	ret.key = argon2.IDKey([]byte(password), ret.salt, 1, 64*1024, 4, 32)
	ret.block, err = aes.NewCipher(ret.key)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// NewServer returns a new server with the given password and
// authorization key. If the password does not validate, returns error
func NewServer(password, authkey string) (Server, error) {
	ret := Server{}
	bauth, err := base64.StdEncoding.DecodeString(authkey)
	if err != nil {
		return ret, err
	}
	if len(bauth) != rndBlockSize+sha256.Size+saltSize+aes.BlockSize {
		return ret, ErrInvalidAuthKey
	}
	ret.salt = bauth[:saltSize]
	ret.iv = bauth[saltSize : saltSize+aes.BlockSize]
	ret.key = argon2.IDKey([]byte(password), ret.salt, 1, 64*1024, 4, 32)
	ret.block, err = aes.NewCipher(ret.key)
	if err != nil {
		return ret, err
	}

	check, err := ret.Decrypt(bauth[saltSize+aes.BlockSize:])
	if err != nil {
		return ret, err
	}

	hmac := make([]byte, sha256.Size)
	sh := sha256.Sum256(check[:rndBlockSize])
	copy(hmac, sh[:])
	if bytes.Compare(hmac, check[rndBlockSize:]) != 0 {
		return ret, ErrInvalidPassword
	}
	return ret, nil
}

// GetAuthKey returns a string that can be used for password validation. It is:
//
//	salt iv enc(randombytes hmac)
func (s Server) GetAuthKey() (string, error) {
	encdata := make([]byte, rndBlockSize+sha256.Size)
	rbytes := encdata[:rndBlockSize]
	hmac := encdata[rndBlockSize:]
	rndb(rbytes)
	x := sha256.Sum256(rbytes)
	copy(hmac, x[:])
	out, err := s.Encrypt(encdata)
	if err != nil {
		return "", err
	}
	ret := append(s.salt, s.iv...)
	ret = append(ret, out...)
	return base64.StdEncoding.EncodeToString(ret), nil
}

// Encrypt the given data block using key
func (s Server) Encrypt(in []byte) ([]byte, error) {
	stream := cipher.NewOFB(s.block, s.iv)
	buf := bytes.Buffer{}
	str := cipher.StreamWriter{S: stream, W: &buf}
	io.Copy(str, bytes.NewReader(in))
	return buf.Bytes(), nil
}

// Decrypt the given data block using key
func (s Server) Decrypt(in []byte) ([]byte, error) {
	stream := cipher.NewOFB(s.block, s.iv)
	str := cipher.StreamReader{S: stream, R: bytes.NewReader(in)}
	out := bytes.Buffer{}
	io.Copy(&out, str)
	return out.Bytes(), nil
}
