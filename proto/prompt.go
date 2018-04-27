package proto

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func AskUsername() (string, error) {
	fmt.Print("User name: ")
	reader := bufio.NewReader(os.Stdin)
	s, e := reader.ReadString('\n')
	if s[len(s)-1] == '\n' {
		return s[:len(s)-1], e
	}
	return s, e
}

func AskPassword() (string, error) {
	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	return string(bytePassword), err
}
