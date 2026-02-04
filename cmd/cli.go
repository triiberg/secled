package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"golang.org/x/term"
)

func readPassword(prompt string) (string, error) {
	b, err := readSecretBytes(prompt)
	if err != nil {
		return "", err
	}
	if len(b) == 0 {
		return "", errors.New("password cannot be empty")
	}
	return string(b), nil
}

func readSecret(prompt string) ([]byte, error) {
	b, err := readSecretBytes(prompt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readSecretBytes(prompt string) ([]byte, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, prompt)
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	reader := bufio.NewReader(os.Stdin)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	data = bytesTrimTrailingNewline(data)
	if len(data) == 0 {
		return nil, errors.New("empty input")
	}
	return data, nil
}

func bytesTrimTrailingNewline(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	if len(b) > 0 && b[len(b)-1] == '\r' {
		b = b[:len(b)-1]
	}
	return b
}

func formatSetEnv(password string) string {
	if runtime.GOOS == "windows" {
		return "$env:SECLED_MASTER=" + quotePowerShell(password)
	}
	return "export SECLED_MASTER=" + quotePOSIX(password)
}

func formatUnsetEnv() string {
	if runtime.GOOS == "windows" {
		return "Remove-Item Env:SECLED_MASTER -ErrorAction SilentlyContinue"
	}
	return "unset SECLED_MASTER"
}

func quotePOSIX(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func quotePowerShell(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func initialValue() string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
	created := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(
		"created_at=%s\nhostname=%s\ngoos=%s\ngoarch=%s\n",
		created,
		host,
		runtime.GOOS,
		runtime.GOARCH,
	)
}
