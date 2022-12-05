//go:build linux

// main is the entrypoint for the command line interface for ipwatch
package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmbaur/ipwatch/ipwatch"
)

func logic() error {
	hooks := []string{}
	ifaces := []string{}
	filters := []string{}

	debug := flag.Bool("debug", false, "Run in debug mode")
	envFile := flag.String("env", "", "File containing environment variables to set when running executable hooks (in the form KEY=VAL on each line)")

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

	watcher, err := ipwatch.NewWatcher(ipwatch.WatcherConfig{Debug: *debug})
	if err != nil {
		return err
	}

	hookEnvironment := []string{}
	if *envFile != "" {
		f, err := os.Open(*envFile)
		if err != nil {
			return err
		}
		contents, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		r := bufio.NewReader(bytes.NewReader(contents))
		for {
			bline, _, err := r.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			line := strings.TrimSpace(string(bline))
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "#") {
				continue
			}
			hookEnvironment = append(hookEnvironment, line)
		}
		f.Close()
	}

	return watcher.Watch(ipwatch.WatchConfig{
		Filters:         filters,
		Hooks:           hooks,
		IPv4:            *doIPv4,
		IPv6:            *doIPv6,
		Interfaces:      ifaces,
		MaxRetries:      *maxRetries,
		HookEnvironment: hookEnvironment,
	})
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
