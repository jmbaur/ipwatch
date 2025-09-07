// Package ipwatch implements everything needed for watching and acting on IP
// address changes to network interfaces.
package ipwatch

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"
	"unsafe"

	systemdDaemon "github.com/coreos/go-systemd/daemon"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

func passesFilters(addr netip.Addr, filters ...string) bool {
	value := reflect.ValueOf(addr)

	for _, filter := range filters {
		var not, result bool

		if strings.HasPrefix(filter, "!") {
			not = true
			filter = filter[1:]
		}

		method := value.MethodByName(filter)
		if !method.IsValid() {
			log.Printf("invalid filter %s, skipping\n", filter)
			continue
		}

		ret := method.Call(nil)
		result = ret[0].Bool()

		if not {
			result = !result
		}

		if !result {
			return false
		}
	}

	return true
}

// Hook is the set of filters and program to run when an IP address on an
// interface changes.
type Hook struct {
	Filters []string
	ipCache map[netip.Addr]struct{}
}

// NewHook makes a new hook.
func NewHook(filters []string) Hook {
	return Hook{
		Filters: filters,
		ipCache: map[netip.Addr]struct{}{},
	}
}

func getIfAddrmsg(ifaddrmsg []byte) (*unix.IfAddrmsg, error) {
	// struct ifaddrmsg {
	//   __u8   ifa_family;
	//   __u8   ifa_prefixlen;  // The prefix length
	//   __u8   ifa_flags;      // Flags
	//   __u8   ifa_scope;      // Address scope
	//   __u32  ifa_index;      // Link index
	// };
	return (*unix.IfAddrmsg)(unsafe.Pointer(&ifaddrmsg[0:unix.SizeofIfAddrmsg][0])), nil
}

func handleDelAddr(msg netlink.Message, hooks map[string]Hook) error {
	log.Println("Handling delete address message")

	ifaddrmsg, err := getIfAddrmsg(msg.Data)
	if err != nil {
		return err
	}

	idx := int(ifaddrmsg.Index)
	gotIface, err := net.InterfaceByIndex(idx)
	if err != nil {
		return fmt.Errorf("failed to get interface by index: %v", err)
	}

	var foundHook *Hook
	for iface, hook := range hooks {
		if iface == gotIface.Name {
			foundHook = &hook
		}
	}

	if foundHook == nil {
		log.Printf("Interface '%s' not found in hooks, skipping\n", gotIface.Name)
		return nil
	}

	ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
	if err != nil {
		return err
	}

	for ad.Next() {
		switch ad.Type() {
		case unix.IFA_ADDRESS:
			{
				ip := ad.Bytes()
				addr, ok := netip.AddrFromSlice(ip)
				if !ok {
					return nil
				}
				log.Println("Deleting address from cache")
				delete(foundHook.ipCache, addr)
			}
		}
	}

	return nil
}

func handleNewAddr(msg netlink.Message, hooks map[string]Hook, startup bool) error {
	log.Println("Handling new address message")

	ifaddrmsg, err := getIfAddrmsg(msg.Data)
	if err != nil {
		return err
	}

	idx := int(ifaddrmsg.Index)
	gotIface, err := net.InterfaceByIndex(idx)
	if err != nil {
		return fmt.Errorf("failed to get interface by index: %v", err)
	}

	var foundHook *Hook
	for iface, hook := range hooks {
		if iface == gotIface.Name {
			foundHook = &hook
		}
	}

	if foundHook == nil {
		log.Printf("Interface '%s' not found in hooks, skipping\n", gotIface.Name)
		return nil
	}

	ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
	if err != nil {
		return err
	}

	var (
		fresh bool
		newIP netip.Addr
	)

	for ad.Next() {
		switch ad.Type() {
		case unix.IFA_CACHEINFO:
			{
				// struct ifa_cacheinfo {
				//   __u32  ifa_prefered;
				//   __u32  ifa_valid;
				//   __u32  cstamp;  // created timestamp, hundredths of seconds
				//   __u32  tstamp;  // updated timestamp, hundredths of seconds
				// };
				ifacacheinfo := (*unix.IfaCacheinfo)(unsafe.Pointer(&ad.Bytes()[0:unix.SizeofIfaCacheinfo][0]))

				var ts syscall.Timespec
				syscall.Syscall(syscall.SYS_CLOCK_GETTIME, unix.CLOCK_BOOTTIME, uintptr(unsafe.Pointer(&ts)), 0)
				boottime := uint64(ts.Sec)
				updatedAt := uint64(ifacacheinfo.Tstamp / 100)
				fresh = time.Duration(boottime-updatedAt)*time.Second < 30*time.Second
			}
		case unix.IFA_ADDRESS:
			{
				ip := ad.Bytes()
				addr, ok := netip.AddrFromSlice(ip)
				if !ok {
					log.Println("Address not of length 4 or 16")
					return nil
				}

				if _, ok := foundHook.ipCache[addr]; ok {
					log.Println("New addr was found in cache, skipping hooks")
					return nil
				}

				newIP = addr
			}
		}
	}

	if !newIP.IsValid() {
		log.Println("No address found, skipping hooks")
		return nil
	}

	if !passesFilters(newIP, foundHook.Filters...) {
		log.Println("Address does not pass filters, skipping hooks")
		return nil
	}

	log.Println("Caching new address", newIP)
	foundHook.ipCache[newIP] = struct{}{}

	if !fresh && startup {
		log.Println("IP address is not new and the program did not just startup, skipping hooks")
		return nil
	} else if fresh && startup {
		log.Println("Fresh IP address and starting up, running hooks")
	}

	out, err := json.Marshal(struct {
		Ifindex   uint32 `json:"ifindex"`
		PrefixLen uint8  `json:"prefixlen"`
		Address   string `json:"address"`
	}{
		Ifindex:   ifaddrmsg.Index,
		PrefixLen: ifaddrmsg.Prefixlen,
		Address:   newIP.String(),
	})
	if err != nil {
		log.Printf("failed to encode json: %v\n", err)
		return err
	}

	if _, err := os.Stdout.Write(out); err != nil {
		return err
	}
	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func generateCache(hooks map[string]Hook) error {
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{Strict: true})
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Println("Caching initial IPs")
	responses, err := conn.Execute(netlink.Message{
		Header: netlink.Header{Type: unix.RTM_GETADDR, Flags: unix.NLM_F_REQUEST | unix.NLM_F_DUMP},
		Data:   (*(*[unix.SizeofIfAddrmsg]byte)(unsafe.Pointer(&unix.IfAddrmsg{Family: unix.AF_UNSPEC})))[:],
	})
	if err != nil {
		return err
	}

	for _, msg := range responses {
		if msg.Header.Type == unix.RTM_NEWADDR {
			if err := handleNewAddr(msg, hooks, true); err != nil {
				return err
			}
		}
	}

	return nil
}

// Watch watches for IP address changes performs hook actions on new IP
// addresses. This function blocks.
func Watch(hooks map[string]Hook) error {
	if len(hooks) == 0 {
		panic("unreachable")
	}

	if err := generateCache(hooks); err != nil {
		return err
	}

	ifaceDisplay := []string{}
	for iface := range hooks {
		ifaceDisplay = append(ifaceDisplay, iface)
	}
	log.Printf(
		"Listening for IP address changes on %s\n",
		strings.Join(ifaceDisplay, ", "),
	)

	notifySupported, err := systemdDaemon.SdNotify(false, systemdDaemon.SdNotifyReady)
	if err != nil {
		return err
	}
	if !notifySupported {
		log.Println("Systemd notify not supported in current running environment")
	}

	log.Println("Opening netlink socket")
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{
		Groups: unix.RTMGRP_IPV4_IFADDR | unix.RTMGRP_IPV6_IFADDR,
		Strict: true,
	})
	if err != nil {
		return fmt.Errorf("failed to dial netlink: %w", err)
	}
	defer conn.Close()

	for {
		msgs, err := conn.Receive()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Received netlink messages")
		for _, msg := range msgs {
			switch msg.Header.Type {
			case unix.RTM_DELADDR:
				if err := handleDelAddr(msg, hooks); err != nil {
					log.Println(err)
				}
			case unix.RTM_NEWADDR:
				if err := handleNewAddr(msg, hooks, false); err != nil {
					log.Println(err)
				}
			}
		}
	}
}
