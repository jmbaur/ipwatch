package cmd

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmbaur/ipwatch/internal"
)

var ErrNoScripts = errors.New("no scripts to run")

func Run() error {
	scripts := []string{}
	ifaces := []string{}

	maxRetries := flag.Int(
		"max-retries",
		3,
		"The maximum number of attempts that will be made for a failing script",
	)
	flag.Func(
		"script",
		"The path to an executable/script to run on IP address change",
		func(script string) error {
			scripts = append(scripts, filepath.Join(script))
			return nil
		},
	)
	flag.Func(
		"interface",
		"The name of an interface to notify for changes",
		func(iface string) error {
			ifaces = append(ifaces, iface)
			return nil
		},
	)
	flag.Parse()

	if len(scripts) == 0 {
		return ErrNoScripts
	}

	if len(ifaces) > 0 {
		fmt.Printf(
			"Listening for IPv4 address changes on %s\n",
			strings.Join(ifaces, ", "),
		)
	} else {
		fmt.Println(
			"Listening for IPv4 address changes on all interfaces",
		)
	}

	if err := internal.Logic(*maxRetries, ifaces, scripts); err != nil {
		return err
	}

	return nil
}
