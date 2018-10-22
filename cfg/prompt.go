package cfg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// Ask asks something to the user and returns it. Panics on error
func Ask(prompt string) string {
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

// AskPasswordWithPrompt prompts, and asks password
func AskPasswordWithPrompt(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	return string(bytePassword)
}
