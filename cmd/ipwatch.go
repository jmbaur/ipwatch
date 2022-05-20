package cmd

import (
	"errors"
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/jmbaur/ipwatch/internal"
)

var ErrNoHookScript = errors.New("Did not provide hook script")

func Run() error {
	exeToRun := flag.String(
		"hook-script",
		"",
		"The path to an executable/script to run on IP address change",
	)
	ifacesFlag := flag.String(
		"interfaces",
		"",
		"A comma-separated list of interfaces to notify for changes",
	)
	flag.Parse()

	if *exeToRun == "" {
		return ErrNoHookScript
	}

	ifacesOfInterest := strings.FieldsFunc(*ifacesFlag, func(r rune) bool {
		return r == ','
	})

	if len(ifacesOfInterest) > 0 {
		log.Printf(
			"listening for IPv4 address changes on %s\n",
			ifacesOfInterest,
		)
	} else {
		log.Println(
			"listening for IPv4 address changes on all interfaces",
		)
	}

	if err := internal.Loop(
		ifacesOfInterest,
		filepath.Join(*exeToRun),
	); err != nil {
		log.Fatal(err)
	}

	return nil
}
