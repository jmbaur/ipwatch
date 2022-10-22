package ipwatch

import (
	"bytes"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"strings"
)

var ErrInvalidHook = errors.New("invalid hook")

type Hook interface {
	Name() string
	// Run returns an output string and an error
	Run(iface string, addr netip.Addr) (string, error)
}

func NewHook(hook string) (Hook, error) {
	split := strings.SplitN(hook, ":", 2)
	if len(split) != 2 {
		return nil, ErrInvalidHook
	}
	hookType := split[0]
	hookName := split[1]
	switch hookType {
	case "internal":
		switch hookName {
		case "echo":
			return &Echo{}, nil
		}
	case "executable":
		return &Executable{ExeName: split[1]}, nil
	}

	return nil, ErrInvalidHook
}

type Echo struct{}

func (e *Echo) Name() string {
	return "internal:echo"
}

func (e *Echo) Run(iface string, addr netip.Addr) (string, error) {
	return fmt.Sprintf("New IP for %s: %s", iface, addr), nil
}

type Executable struct {
	// The name or full path to an executable
	ExeName string
}

func (e *Executable) Name() string {
	return fmt.Sprintf("executable:%s", e.ExeName)
}

func (e *Executable) Run(iface string, addr netip.Addr) (string, error) {
	cmd := exec.Command(e.ExeName)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("IFACE=%s", iface))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ADDR=%s", addr))

	output, err := cmd.CombinedOutput()
	return string(bytes.TrimSpace(output)), err
}
