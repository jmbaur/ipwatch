package ipwatch

import (
	"bytes"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
)

func runHook(hookPath string, ifaceIdx uint32, addr netip.Addr) (string, error) {
	cmd := exec.Command(hookPath)
	cmd.Env = append(cmd.Env, os.Environ()...) // inherit from current environment
	cmd.Env = append(cmd.Env, fmt.Sprintf("IFACE_IDX=%d", ifaceIdx))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ADDR=%s", addr))

	switch {
	case addr.Is6():
		cmd.Env = append(cmd.Env, "IS_IP6=1")
	case addr.Is4():
		cmd.Env = append(cmd.Env, "IS_IP4=1")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(output)), nil
}
