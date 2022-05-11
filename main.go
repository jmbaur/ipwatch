package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

func main() {
	exeToRun := flag.String("exe", "", "The path to an executable to run")
	ifaceName := flag.String("iface", "", "The interface to listen for changes on")
	flag.Parse()

	if *exeToRun == "" {
		flag.Usage()
		os.Exit(1)
	}

	c, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{
		Groups: unix.RTMGRP_IPV4_IFADDR,
	})
	if err != nil {
		log.Fatalf("failed to dial netlink: %v", err)
	}
	defer c.Close()

	log.Println("listening for IPv4 address changes")

	for {
		msgs, err := c.Receive()
		if err != nil {
			log.Println(err)
			break
		}
		for _, msg := range msgs {
			if msg.Header.Type == unix.RTM_NEWADDR {
				iface, err := net.InterfaceByIndex(int(msg.Data[4]))
				if err != nil {
					log.Println(err)
					continue
				}
				if *ifaceName != "" && iface.Name != *ifaceName {
					continue
				}

				ip := net.IPv4(msg.Data[12], msg.Data[13], msg.Data[14], msg.Data[15])

				cmd := exec.Command(filepath.Join(*exeToRun))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Env = append(cmd.Env, os.Environ()...)
				cmd.Env = append(cmd.Env, fmt.Sprintf("IFACE=%s", iface.Name))
				cmd.Env = append(cmd.Env, fmt.Sprintf("ADDR=%s", ip))
				log.Println("================================================================================")
				if err := cmd.Run(); err != nil {
					log.Printf("Error running exe: %v", err)
				}
				log.Println("================================================================================")
			}
		}
	}
}
