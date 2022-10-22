package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/jmbaur/ipwatch/ipwatch"
)

func logic() error {
	scripts := []string{}
	ifaces := []string{}
	filters := []string{}

	maxRetries := flag.Uint(
		"max-retries",
		3,
		"The maximum number of attempts that will be made for a failing script.",
	)

	doIPv4 := flag.Bool("4", false, "Watch only for IPv4 address changes")
	doIPv6 := flag.Bool("6", false, "Watch only for IPv6 address changes")

	flag.Func("filter",
		"Conditions that must be true before running scripts. See methods for https://pkg.go.dev/net/netip#Addr that start with 'Is'.",
		func(filter string) error {
			filters = append(filters, filter)
			return nil
		})

	flag.Func(
		"script",
		"The path to an executable/script to run on IP address change.",
		func(script string) error {
			scripts = append(scripts, filepath.Join(script))
			return nil
		},
	)
	flag.Func(
		"interface",
		"The name of an interface to notify for changes.",
		func(iface string) error {
			ifaces = append(ifaces, iface)
			return nil
		},
	)
	flag.Parse()

	if len(ifaces) > 0 {
		fmt.Printf(
			"Listening for IP address changes on %s\n",
			strings.Join(ifaces, ", "),
		)
	} else {
		fmt.Println(
			"Listening for IP address changes on all interfaces",
		)
	}

	if watcher, err := ipwatch.NewWatcher(ipwatch.WatchConfig{
		MaxRetries: *maxRetries,
		Interfaces: ifaces,
		Scripts:    scripts,
		IPv4:       *doIPv4,
		IPv6:       *doIPv6,
		Filters:    filters,
	}); err != nil {
		return err
	} else {
		return watcher.Watch()
	}
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
