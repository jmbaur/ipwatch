//go:build linux

// main is the entrypoint for the command line interface for ipwatch
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/jmbaur/ipwatch/ipwatch"
)

func logic() error {
	hooks := map[string]ipwatch.Hook{}

	flag.Func(
		"hook",
		"Hook specifier of the form <iface>:<filter1>,<filter2>,...",
		func(hookStr string) error {
			split := strings.SplitN(hookStr, ":", 2)
			if len(split) != 2 {
				return fmt.Errorf("invalid hook specifier %s", hookStr)
			}

			if _, duplicate := hooks[split[0]]; duplicate {
				return fmt.Errorf("hook for interface %s specified more than once", split[0])
			}

			filters := strings.Split(split[1], ",")
			if filters[0] == "" {
				filters = []string{}
			}

			hooks[split[0]] = ipwatch.NewHook(
				filters,
			)

			return nil
		},
	)
	flag.Parse()

	if len(hooks) == 0 {
		return fmt.Errorf("no hooks specified")
	}

	return ipwatch.Watch(hooks)
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
