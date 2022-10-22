package ipwatch

import (
	"bytes"
	"fmt"
	"math"
	"net/netip"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

type WatchConfig struct {
	MaxRetries uint
	Interfaces []string
	Scripts    []string
	IPv4       bool
	IPv6       bool
}

func (cfg *WatchConfig) Watch() error {
	var groups uint32 = 0
	if cfg.IPv4 {
		groups |= unix.RTMGRP_IPV4_IFADDR
	}
	if cfg.IPv6 {
		groups |= unix.RTMGRP_IPV6_IFADDR
	}

	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{
		Groups: groups,
		Strict: true,
	})
	if err != nil {
		return fmt.Errorf("failed to dial netlink: %v", err)
	}
	defer conn.Close()

	for {
		msgs, err := conn.Receive()
		if err != nil {
			fmt.Printf("Failed to receive messages: %v", err)
			break
		}
	messages:
		for _, msg := range msgs {
			if msg.Header.Type != unix.RTM_NEWADDR {
				continue
			}

			var ifaceName string
			var newIP netip.Addr

			ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
			if err != nil {
				fmt.Printf("Could not get attribute decoder: %v", err)
				continue
			}

			for ad.Next() {
				switch ad.Type() {
				case unix.IFA_ADDRESS:
					ip := ad.Bytes()
					var ok bool
					newIP, ok = netip.AddrFromSlice(ip)
					if !ok {
						continue messages
					}
				case unix.IFA_LABEL:
					ifaceName = ad.String()
					interested := len(cfg.Interfaces) == 0
					for _, iface := range cfg.Interfaces {
						interested = ifaceName == iface
					}
					if !interested {
						continue messages
					}
				}
			}

			var wg sync.WaitGroup
			var l sync.Mutex

			for _, script := range cfg.Scripts {
				wg.Add(1)
				script := script
				go func() {
					for i := 1; i <= int(cfg.MaxRetries); i++ {
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
							if i < int(cfg.MaxRetries) {
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
	var largest int
	for _, output := range outputs {
		for _, line := range strings.Split(output, "\n") {
			if len(line) > largest {
				largest = len(line)
			}
		}
	}

	separator := strings.Repeat("-", largest)
	fmt.Println(separator)
	for _, output := range outputs {
		fmt.Println(output)
	}
	fmt.Println(separator)
}
