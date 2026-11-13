// Package creds stores the Buttons login in the OS keychain (macOS Keychain,
// Windows Credential Manager, or the Secret Service on Linux).
package creds

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const service = "buttons-mcp"

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Load returns credentials for url: BUTTONS_USERNAME/BUTTONS_PASSWORD if set,
// otherwise the keychain entry saved by `buttons-mcp login`.
func Load(url string) (Login, error) {
	if u := os.Getenv("BUTTONS_USERNAME"); u != "" {
		return Login{u, os.Getenv("BUTTONS_PASSWORD")}, nil
	}
	s, err := keyring.Get(service, url)
	if err != nil {
		return Login{}, fmt.Errorf("no Buttons login saved for %s; run `buttons-mcp login`", url)
	}
	var l Login
	err = json.Unmarshal([]byte(s), &l)
	return l, err
}

func Save(url string, l Login) error {
	b, _ := json.Marshal(l)
	return keyring.Set(service, url, string(b))
}

func Delete(url string) error { return keyring.Delete(service, url) }

// Prompt asks for username and password, in the terminal when there is one,
// otherwise in a macOS dialog.
func Prompt() (Login, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "Buttons username: ")
		u, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return Login{}, err
		}
		fmt.Fprint(os.Stderr, "Buttons password: ")
		p, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return Login{}, err
		}
		return Login{strings.TrimSpace(u), string(p)}, nil
	}
	if runtime.GOOS == "darwin" {
		u, err := dialog("Bitfocus Buttons username:", false)
		if err != nil {
			return Login{}, err
		}
		p, err := dialog("Bitfocus Buttons password for "+u+":", true)
		return Login{u, p}, err
	}
	return Login{}, errors.New("no terminal to prompt in; set BUTTONS_USERNAME and BUTTONS_PASSWORD")
}

func dialog(prompt string, hidden bool) (string, error) {
	script := `text returned of (display dialog "` + prompt + `" default answer "" with title "buttons-mcp"`
	if hidden {
		script += " with hidden answer"
	}
	script += ")"
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", errors.New("login cancelled")
	}
	return strings.TrimRight(string(out), "\n"), nil
}
