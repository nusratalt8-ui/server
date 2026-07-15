package logger

import (
	"bufio"
	"os"
	"strings"

	"golang.org/x/term"
)

var reader = bufio.NewReader(os.Stdin)

func Prompt(label string) string {
	mu.Lock()
	os.Stdout.WriteString(cyan + bold + "? " + reset + label + " ")
	mu.Unlock()
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func Password(label string) string {
	mu.Lock()
	os.Stdout.WriteString(cyan + bold + "? " + reset + label + " ")
	mu.Unlock()
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	os.Stdout.WriteString("\n")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
