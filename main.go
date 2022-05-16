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
	ifaceOfInterest := flag.String("iface", "", "The interface to listen for changes on")
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
			if msg.Header.Type == unix.NLMSG_DONE {
				continue
			}
			if msg.Header.Type == unix.NLMSG_ERROR {
				// TODO(jared): parse error
				log.Println("NLMSG_ERROR", msg.Data)
				continue
			}
			if msg.Header.Type == unix.RTM_NEWADDR {
				var ifaceName, newIP string

				decoder, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
				if err != nil {
					log.Println(err)
					continue
				}

				for decoder.Next() {
					switch decoder.Type() {
					case unix.IFA_ADDRESS:
						ip := decoder.Bytes()
						if len(ip) != 4 {
							log.Println("Did not get correct number of bytes")
							continue
						}
						newIP = net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
					case unix.IFA_LABEL:
						ifaceName = decoder.String()
						if *ifaceOfInterest != "" && ifaceName != *ifaceOfInterest {
							continue
						}
					}
				}

				cmd := exec.Command(filepath.Join(*exeToRun))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Env = append(cmd.Env, os.Environ()...)
				cmd.Env = append(cmd.Env, fmt.Sprintf("IFACE=%s", ifaceName))
				cmd.Env = append(cmd.Env, fmt.Sprintf("ADDR=%s", newIP))
				log.Println("================================================================================")
				if err := cmd.Run(); err != nil {
					log.Printf("Error running exe: %v", err)
				}
				log.Println("================================================================================")
			}
		}
	}
}
