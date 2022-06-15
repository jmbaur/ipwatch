package internal

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

var separator = strings.Repeat("=", 80)

func getCache(ifaces []string) (map[string]net.IP, error) {
	cache := make(map[string]net.IP)

	for _, iface := range ifaces {
		ifi, err := net.InterfaceByName(iface)
		if err != nil {
			return nil, err
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			return nil, err
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

	return cache, nil
}

func Logic(maxRetries int, ifaces []string, scripts []string) error {
	cache, err := getCache(ifaces)
	if err != nil {
		return err
	}

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
					interested := len(ifaces) == 0
					for _, iface := range ifaces {
						interested = ifaceName == iface
					}
					if !interested {
						continue messages
					}
				}
			}

			if oldIP, ok := cache[ifaceName]; ok && oldIP.String() == newIP.String() {
				log.Printf("New IP for %s has not changed, not calling scripts\n", ifaceName)
				continue
			}

			cache[ifaceName] = newIP

			var wg sync.WaitGroup
			var l sync.Mutex

			for _, script := range scripts {
				wg.Add(1)
				script := script
				go func() {
					for i := 1; i <= maxRetries; i++ {
						backoff := time.Duration(math.Pow(2, float64(i))) * time.Second

						cmd := exec.Command(script)
						cmd.Env = append(cmd.Env, os.Environ()...)
						cmd.Env = append(cmd.Env, fmt.Sprintf("IFACE=%s", ifaceName))
						cmd.Env = append(cmd.Env, fmt.Sprintf("ADDR=%s", newIP))

						if output, err := cmd.CombinedOutput(); err != nil {
							outputs := []string{}
							outputs = append(outputs, fmt.Sprintf("Error running script %s", script))
							trimmedOutput := bytes.TrimSpace(output)
							if len(trimmedOutput) > 0 {
								outputs = append(outputs, string(trimmedOutput))
							}
							outputs = append(outputs, err.Error())
							if i < maxRetries {
								outputs = append(outputs, fmt.Sprintf("Retrying in %s", backoff))
							} else {
								outputs = append(outputs, "Max attempts reached")
							}

							l.Lock()
							printOutput(outputs...)
							l.Unlock()

							time.Sleep(backoff)
							continue
						} else {
							outputs := []string{}
							outputs = append(outputs, fmt.Sprintf("Script %s succeeded", script))
							trimmedOutput := bytes.TrimSpace(output)
							if len(trimmedOutput) > 0 {
								outputs = append(outputs, string(trimmedOutput))
							}

							l.Lock()
							printOutput(outputs...)
							l.Unlock()

							wg.Done()
							break
						}
					}
				}()
			}

			wg.Wait()
		}
	}

	return nil
}

func printOutput(outputs ...string) {
	log.Println(separator)
	for _, output := range outputs {
		log.Println(output)
	}
	log.Println(separator)
}
