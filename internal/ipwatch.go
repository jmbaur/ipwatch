package internal

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

func getCache(ifaces []string) map[string]net.IP {
	cache := make(map[string]net.IP)

	for _, iface := range ifaces {
		ifi, err := net.InterfaceByName(iface)
		if err != nil {
			log.Println(err)
			continue
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			v4Addr := ipnet.IP.To4()
			if v4Addr == nil {
				continue
			}
			cache[iface] = v4Addr
		}
	}

	return cache
}

func Loop(ifaces []string, exe string) error {
	cache := getCache(ifaces)

	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{
		Groups: unix.RTMGRP_IPV4_IFADDR,
		Strict: true,
	})
	if err != nil {
		return fmt.Errorf("failed to dial netlink: %v", err)
	}
	defer conn.Close()

	for {
		msgs, err := conn.Receive()
		if err != nil {
			log.Printf("Failed to receive messages: %v", err)
			break
		}
	messages:
		for _, msg := range msgs {
			if msg.Header.Type != unix.RTM_NEWADDR {
				continue
			}

			var ifaceName string
			var newIP net.IP

			ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
			if err != nil {
				log.Printf("Could not get attribute decoder: %v", err)
				continue
			}

			for ad.Next() {
				switch ad.Type() {
				case unix.IFA_ADDRESS:
					ip := ad.Bytes()
					if len(ip) != 4 {
						log.Println("Did not get correct number of bytes")
						continue
					}
					newIP = net.IPv4(ip[0], ip[1], ip[2], ip[3])
				case unix.IFA_LABEL:
					ifaceName = ad.String()
					interested := false
					for _, iface := range ifaces {
						interested = ifaceName == iface
					}
					if !interested {
						continue messages
					}
				}
			}

			oldIP, ok := cache[ifaceName]
			if ok && oldIP.String() == newIP.String() {
				log.Printf("New IP for %s has not changed, not calling hook script\n", ifaceName)
				continue
			}

			cache[ifaceName] = newIP
			cmd := exec.Command(exe)
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

	return nil
}
