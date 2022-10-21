package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/jmbaur/ipwatch/ipwatch"
)

var (
	ErrNoScripts         = errors.New("no scripts to run")
	ErrInvalidIpProtocol = errors.New("only one of -4 and -6 allowed")
	ErrNotImplemented    = errors.New("not implemented")
)

func logic() error {
	scripts := []string{}
	ifaces := []string{}

	ipv4Only := flag.Bool("4", false, "Watch only for IPv4 address changes")
	ipv6Only := flag.Bool("6", false, "Watch only for IPv6 address changes")

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

	if *ipv4Only && *ipv6Only {
		return ErrInvalidIpProtocol
	}

	if *ipv6Only {
		return fmt.Errorf("ipv6: %w", ErrNotImplemented)
	}

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

	if err := ipwatch.Watch(*maxRetries, ifaces, scripts); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
