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

// ErrInvalidHook indicates that the hook type is not supported.
var ErrInvalidHook = errors.New("invalid hook")

// Hook is the interface required for running an arbitrary step after an IP
// address change.
type Hook interface {
	Name() string
	Run(ifaceIdx uint32, addr netip.Addr) (string, error)
}

// NewHook returns a hook, parsing the hook format '<hook name>:<hook value>'.
func NewHook(hook string, hookEnvironment []string) (Hook, error) {
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
		return &Executable{ExeName: split[1], Environment: hookEnvironment}, nil
	}

	return nil, ErrInvalidHook
}

// Echo is a hook that simply prints new IP address information to the screen.
// To use Echo, provide the hook with the format 'internal:echo'.
type Echo struct{}

// Name implements Hook
func (e *Echo) Name() string {
	return "internal:echo"
}

// Run implements Hook
func (e *Echo) Run(ifaceIdx uint32, addr netip.Addr) (string, error) {
	return fmt.Sprintf("New IP for interface index %d: %s", ifaceIdx, addr), nil
}

// Executable is a hook that can run an arbitrary executable after an IP
// address change. To use this hook, provide the hook format 'executable:<path
// to executable>'.
type Executable struct {
	// The name or full path to an executable
	ExeName string
	// Environment contains strings representing the environment, in the form
	// "key=value" (the same as os.Environ).
	Environment []string
}

// Name implements Hook
func (e *Executable) Name() string {
	return fmt.Sprintf("executable:%s", e.ExeName)
}

// Run implements Hook
func (e *Executable) Run(ifaceIdx uint32, addr netip.Addr) (string, error) {
	cmd := exec.Command(e.ExeName)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, e.Environment...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("IFACE_IDX=%d", ifaceIdx))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ADDR=%s", addr))

	if addr.Is6() {
		cmd.Env = append(cmd.Env, "IS_IP6=1")
	} else if addr.Is4() {
		cmd.Env = append(cmd.Env, "IS_IP4=1")
	}

	output, err := cmd.CombinedOutput()
	return string(bytes.TrimSpace(output)), err
}
