// main is the entrypoint for the command line interface for ipwatch
package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/jmbaur/ipwatch/ipwatch"
)

func logic() error {
	hooks := []string{}
	ifaces := []string{}
	filters := []string{}

	debug := flag.Bool("debug", false, "Run in debug mode")

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
		"hook",
		"A hook to run upon receiving a new IP address.",
		func(script string) error {
			hooks = append(hooks, filepath.Join(script))
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

	watcher, err := ipwatch.NewWatcher(ipwatch.WatchConfig{
		Debug:      *debug,
		MaxRetries: *maxRetries,
		Interfaces: ifaces,
		Hooks:      hooks,
		IPv4:       *doIPv4,
		IPv6:       *doIPv6,
		Filters:    filters,
	})
	if err != nil {
		return err
	}

	return watcher.Watch()
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
