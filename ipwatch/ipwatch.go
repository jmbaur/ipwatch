// Package ipwatch implements everything needed for watching and acting on IP
// address changes to network interfaces.
package ipwatch

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"strings"
	"syscall"
	"time"
	"unsafe"

	systemdDaemon "github.com/coreos/go-systemd/daemon"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

var (
	// ErrIncorrectSizeOf indicates the incorrect size was present
	ErrIncorrectSizeOf = errors.New("incorrect size of")
	// ErrInvalidFilter indicates that the filter is not supported
	ErrInvalidFilter = errors.New("invalid filter")
	// ErrInvalidIPProtocol indicates that the IP protocol is not supported
	ErrInvalidIPProtocol = errors.New("only one of -4 and -6 allowed")
)

var validFilters = map[string]struct{}{
	"Is4":                       {},
	"Is4In6":                    {},
	"Is6":                       {},
	"IsGlobalUnicast":           {},
	"IsInterfaceLocalMulticast": {},
	"IsLinkLocalMulticast":      {},
	"IsLinkLocalUnicast":        {},
	"IsLoopback":                {},
	"IsMulticast":               {},
	"IsPrivate":                 {},
	"IsUnspecified":             {},
	"IsValid":                   {},
}

func printOutput(outputs ...string) {
	separator := strings.Repeat("=", 3)
	fmt.Println(separator)
	for _, output := range outputs {
		fmt.Println(output)
	}
	fmt.Println(separator)
}

func isValidFilter(filter string) bool {
	test := filter
	if strings.HasPrefix(filter, "!") {
		test = filter[1:]
	}
	_, ok := validFilters[test]
	return ok
}

func passesFilters(addr netip.Addr, filters ...string) bool {
	for _, filter := range filters {
		var not, result bool

		if strings.HasPrefix(filter, "!") {
			not = true
			filter = filter[1:]
		}
		switch filter {
		case "Is4":
			result = addr.Is4()
		case "Is4In6":
			result = addr.Is4In6()
		case "Is6":
			result = addr.Is6()
		case "IsGlobalUnicast":
			result = addr.IsGlobalUnicast()
		case "IsInterfaceLocalMulticast":
			result = addr.IsInterfaceLocalMulticast()
		case "IsLinkLocalMulticast":
			result = addr.IsLinkLocalMulticast()
		case "IsLinkLocalUnicast":
			result = addr.IsLinkLocalUnicast()
		case "IsLoopback":
			result = addr.IsLoopback()
		case "IsMulticast":
			result = addr.IsMulticast()
		case "IsPrivate":
			result = addr.IsPrivate()
		case "IsUnspecified":
			result = addr.IsUnspecified()
		case "IsValid":
			result = addr.IsValid()
		}

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
	Program string
	Filters []string
	ipCache map[netip.Addr]struct{}
}

// NewHook makes a new hook.
func NewHook(program string, filters []string) Hook {
	return Hook{
		Program: program,
		Filters: filters,
		ipCache: map[netip.Addr]struct{}{},
	}
}

// WatchConfig sets filters and hooks for the watcher.
type WatchConfig struct {
	Hooks      []string
	Interfaces []string
	MaxRetries uint
}

// Watcher can be used to watch changes to IP addresses, optionally filtering
// on interface and/or IP address types.
type Watcher struct {
	log *log.Logger
}

// NewWatcher parses a WatchConfig and returns a new Watcher.
func NewWatcher() (*Watcher, error) {
	return &Watcher{log: log.Default()}, nil
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

func (w *Watcher) handleDelAddr(msg netlink.Message, hooks map[string]Hook) error {
	w.log.Println("Handling delete address message")

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
		w.log.Printf("Interface '%s' not found in hooks, skipping\n", gotIface.Name)
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
				w.log.Println("Deleting address from cache")
				delete(foundHook.ipCache, addr)
			}
		}
	}

	return nil
}

func (w *Watcher) handleNewAddr(msg netlink.Message, hooks map[string]Hook, startup bool) error {
	w.log.Println("Handling new address message")

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
		w.log.Printf("Interface '%s' not found in hooks, skipping\n", gotIface.Name)
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
					w.log.Println("Address not of length 4 or 16")
					return nil
				}

				if _, ok := foundHook.ipCache[addr]; ok {
					w.log.Println("New addr was found in cache, skipping hooks")
					return nil
				}

				newIP = addr
			}
		}
	}

	if !newIP.IsValid() {
		w.log.Println("No address found, skipping hooks")
		return nil
	}

	if !passesFilters(newIP, foundHook.Filters...) {
		w.log.Println("Address does not pass filters, skipping hooks")
		return nil
	}

	w.log.Println("Caching new address", newIP)
	foundHook.ipCache[newIP] = struct{}{}

	if !fresh && startup {
		w.log.Println("IP address is not new and the program did not just startup, skipping hooks")
		return nil
	} else if fresh && startup {
		w.log.Println("Fresh IP address and starting up, running hooks")
	}

	{
		outputs := []string{}

		program := foundHook.Program
		hookOutput, err := runHook(program, ifaddrmsg.Index, newIP)
		if err == nil {
			outputs = append(outputs, fmt.Sprintf("Hook '%s' succeeded", program))
		} else {
			outputs = append(outputs, fmt.Sprintf("Hook '%s' failed", program))
			outputs = append(outputs, err.Error())
		}

		if len(hookOutput) > 0 {
			outputs = append(outputs, hookOutput)
		}

		printOutput(outputs...)
	}

	return nil
}

func (w *Watcher) generateCache(hooks map[string]Hook) error {
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{Strict: true})
	if err != nil {
		return err
	}
	defer conn.Close()

	w.log.Println("Caching initial IPs")
	responses, err := conn.Execute(netlink.Message{
		Header: netlink.Header{Type: unix.RTM_GETADDR, Flags: unix.NLM_F_REQUEST | unix.NLM_F_DUMP},
		Data:   (*(*[unix.SizeofIfAddrmsg]byte)(unsafe.Pointer(&unix.IfAddrmsg{Family: unix.AF_UNSPEC})))[:],
	})
	if err != nil {
		return err
	}

	for _, msg := range responses {
		if msg.Header.Type == unix.RTM_NEWADDR {
			if err := w.handleNewAddr(msg, hooks, true); err != nil {
				return err
			}
		}
	}

	return nil
}

// Watch watches for IP address changes performs hook actions on new IP
// addresses. This function blocks.
func (w *Watcher) Watch(hooks map[string]Hook) error {
	if len(hooks) == 0 {
		panic("unreachable")
	}

	filterMap := map[string]struct{}{}
	for _, hook := range hooks {
		for _, filter := range hook.Filters {
			if !isValidFilter(filter) {
				return fmt.Errorf("%w: %s", ErrInvalidFilter, filter)
			}
			filterMap[filter] = struct{}{}
		}

		filtersDedup := []string{}
		for k := range filterMap {
			filtersDedup = append(filtersDedup, k)
		}

		hook.Filters = filtersDedup
	}

	if err := w.generateCache(hooks); err != nil {
		return err
	}

	ifaceDisplay := []string{}
	for iface := range hooks {
		ifaceDisplay = append(ifaceDisplay, iface)
	}
	w.log.Printf(
		"Listening for IP address changes on %s\n",
		strings.Join(ifaceDisplay, ", "),
	)

	notifySupported, err := systemdDaemon.SdNotify(false, systemdDaemon.SdNotifyReady)
	if err != nil {
		return err
	}
	if !notifySupported {
		w.log.Println("Systemd notify not supported in current running environment")
	}

	w.log.Println("Opening netlink socket")
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
			w.log.Println(err)
			continue
		}
		w.log.Println("Received netlink messages")
		for _, msg := range msgs {
			switch msg.Header.Type {
			case unix.RTM_DELADDR:
				if err := w.handleDelAddr(msg, hooks); err != nil {
					w.log.Println(err)
				}
			case unix.RTM_NEWADDR:
				if err := w.handleNewAddr(msg, hooks, false); err != nil {
					w.log.Println(err)
				}
			}
		}
	}
}
