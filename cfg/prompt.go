package cfg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var Ask = DefaultAsk
var AskPasswordWithPrompt = DefaultAskPasswordWithPrompt

// DefaultAsk asks something to the user and returns it. Panics on error
func DefaultAsk(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	s, e := reader.ReadString('\n')
	if e != nil {
		log.Fatal(e)
	}
	if s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

// AskUsername asks "user name" and reads input
func AskUsername() string {
	return Ask("User name: ")
}

// AskPassword asks "Password" and reads the password
func AskPassword() string {
	return AskPasswordWithPrompt("Password: ")
}

// DefaultAskPasswordWithPrompt prompts, and asks password
func DefaultAskPasswordWithPrompt(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	return string(bytePassword)
}
