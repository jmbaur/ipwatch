// Package ipwatch implements everything needed for watching and acting on IP
// address changes to network interfaces.
package ipwatch

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/netip"
	"strings"
	"sync"
	"time"
	"unsafe"

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

func isValidFilter(filter string) bool {
	test := filter
	if strings.HasPrefix(filter, "!") {
		test = filter[1:]
	}
	_, ok := validFilters[test]
	return ok
}

func passesFilter(addr netip.Addr, filters ...string) bool {
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

// WatchConfig is used to create a watcher
type WatchConfig struct {
	Debug      bool
	MaxRetries uint
	Interfaces []string
	Hooks      []string
	IPv4       bool
	IPv6       bool
	Filters    []string
}

// Watcher can be used to watch changes to IP addresses, optionally filtering
// on interface and/or IP address types.
type Watcher struct {
	log        *log.Logger
	interfaces []string
	hooks      []Hook
	filters    []string
	maxRetries uint

	l       sync.Mutex
	ipCache map[int]map[netip.Addr]struct{}
}

// NewWatcher parses a WatchConfig and returns a new Watcher.
func NewWatcher(cfg WatchConfig) (*Watcher, error) {
	if cfg.IPv4 && cfg.IPv6 {
		return nil, ErrInvalidIPProtocol
	}

	if cfg.IPv4 {
		cfg.Filters = append(cfg.Filters, "Is4")
	}
	if cfg.IPv6 {
		cfg.Filters = append(cfg.Filters, "Is6")
	}

	if len(cfg.Hooks) == 0 {
		cfg.Hooks = append(cfg.Hooks, "internal:echo")
	}
	hooks := []Hook{}
	for _, hookName := range cfg.Hooks {
		hook, err := NewHook(hookName)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, hookName)
		}
		hooks = append(hooks, hook)
	}

	filterMap := map[string]struct{}{}
	for _, filter := range cfg.Filters {
		if !isValidFilter(filter) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFilter, filter)
		}
		filterMap[filter] = struct{}{}
	}

	filters := []string{}
	for k := range filterMap {
		filters = append(filters, k)
	}

	var logger *log.Logger
	if cfg.Debug {
		logger = log.Default()
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	return &Watcher{
		log:        logger,
		interfaces: cfg.Interfaces,
		hooks:      hooks,
		filters:    filters,
		maxRetries: cfg.MaxRetries,
		ipCache:    map[int]map[netip.Addr]struct{}{},
	}, nil
}

func getIfInfomsg(ifinfomsg []byte) (*unix.IfInfomsg, error) {
	if len(ifinfomsg) < unix.SizeofIfInfomsg {
		return nil, fmt.Errorf("%w: %s", ErrIncorrectSizeOf, "ifinfomsg")
	}

	return (*unix.IfInfomsg)(unsafe.Pointer(&ifinfomsg[0:unix.SizeofIfInfomsg][0])), nil
}

func getIfAddrmsg(ifaddrmsg []byte) (*unix.IfAddrmsg, error) {
	if len(ifaddrmsg) < unix.SizeofIfAddrmsg {
		return nil, fmt.Errorf("%w: %s", ErrIncorrectSizeOf, "ifaddrmsg")
	}

	return (*unix.IfAddrmsg)(unsafe.Pointer(&ifaddrmsg[0:unix.SizeofIfAddrmsg][0])), nil
}

func (w *Watcher) handleDelAddr(msg netlink.Message) error {
	w.log.Println("Handling delete address message")

	ifaddrmsg, err := getIfAddrmsg(msg.Data)
	if err != nil {
		return err
	}

	idx := int(ifaddrmsg.Index)
	if _, ok := w.ipCache[idx]; !ok {
		w.log.Printf("Interface index %d not found in cache, skipping\n", idx)
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
				w.l.Lock()
				if _, ok := w.ipCache[idx]; ok {
					delete(w.ipCache[idx], addr)
				}
				w.l.Unlock()
			}
		}
	}

	return nil
}

func (w *Watcher) handleNewAddr(msg netlink.Message, runHooks bool) error {
	w.log.Println("Handling new address message")

	ifaddrmsg, err := getIfAddrmsg(msg.Data)
	if err != nil {
		return err
	}

	idx := int(ifaddrmsg.Index)
	if _, ok := w.ipCache[idx]; !ok {
		w.log.Printf("Interface index %d not found in cache, skipping\n", idx)
		return nil
	}

	ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
	if err != nil {
		return err
	}

	var newIP *netip.Addr
	for ad.Next() {
		if ad.Type() != unix.IFA_ADDRESS {
			continue
		}
		ip := ad.Bytes()
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			w.log.Println("Address not of length 4 or 16")
			return nil
		}

		if cachedIface, ok := w.ipCache[idx]; ok {
			if _, ok := cachedIface[addr]; ok {
				w.log.Println("New addr was found in cache, skipping hooks")
				return nil
			}
		}

		newIP = &addr
		if !passesFilter(*newIP, w.filters...) {
			w.log.Println("Address does not pass filters, skipping hooks")
			return nil
		}

		w.log.Println("Caching new address")
		w.cacheAddr(idx, addr)
	}

	if !runHooks {
		w.log.Println("runHooks set to false, skipping hooks")
		return nil
	}

	if newIP == nil {
		w.log.Println("No address found, skipping hooks")
		return nil
	}

	var wg sync.WaitGroup
	var l sync.Mutex

	for _, hook := range w.hooks {
		wg.Add(1)
		hook := hook
		go func() {
			for i := 1; i <= int(w.maxRetries); i++ {
				backoff := time.Duration(math.Pow(2, float64(i))) * time.Second

				if hookOutput, err := hook.Run(ifaddrmsg.Index, *newIP); err != nil {
					outputs := []string{}
					outputs = append(outputs, fmt.Sprintf("Hook '%s' failed", hook.Name()))
					if len(hookOutput) > 0 {
						outputs = append(outputs, hookOutput)
					}
					outputs = append(outputs, err.Error())
					if i < int(w.maxRetries) {
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
					outputs = append(outputs, fmt.Sprintf("Hook '%s' succeeded", hook.Name()))
					if len(hookOutput) > 0 {
						outputs = append(outputs, hookOutput)
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

	return nil
}

func (w *Watcher) cacheAddr(iface int, addr netip.Addr) {
	w.l.Lock()
	defer w.l.Unlock()
	if _, ok := w.ipCache[iface]; ok {
		w.ipCache[iface][addr] = struct{}{}
	}
}

func (w *Watcher) generateCache() error {
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{Strict: true})
	if err != nil {
		return err
	}
	defer conn.Close()

	w.log.Println("Hydrating cache with network interfaces")
	linkResponses, err := conn.Execute(netlink.Message{
		Header: netlink.Header{Type: unix.RTM_GETLINK, Flags: unix.NLM_F_REQUEST | unix.NLM_F_DUMP},
		Data:   (*(*[unix.SizeofIfInfomsg]byte)(unsafe.Pointer(&unix.IfInfomsg{Family: unix.AF_UNSPEC})))[:],
	})
	if err != nil {
		return err
	}
	for _, msg := range linkResponses {
		if msg.Header.Type != unix.RTM_NEWLINK {
			continue
		}
		ifinfomsg, err := getIfInfomsg(msg.Data)
		if err != nil {
			return err
		}
		ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfInfomsg:])
		if err != nil {
			return err
		}
		for ad.Next() {
			switch ad.Type() {
			case unix.IFA_LABEL:
				if len(w.interfaces) > 0 {
					for _, iface := range w.interfaces {
						if iface == ad.String() {
							w.l.Lock()
							w.ipCache[int(ifinfomsg.Index)] = map[netip.Addr]struct{}{}
							w.l.Unlock()
						}
					}
				} else {
					w.l.Lock()
					w.ipCache[int(ifinfomsg.Index)] = map[netip.Addr]struct{}{}
					w.l.Unlock()
				}
			}
		}
	}

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
			runHooks := false
			if err := w.handleNewAddr(msg, runHooks); err != nil {
				return err
			}
		}
	}

	return nil
}

// Watch watches for IP address changes performs hook actions on new IP
// addresses. This function blocks.
func (w *Watcher) Watch() error {
	if err := w.generateCache(); err != nil {
		return err
	}

	if len(w.interfaces) > 0 {
		w.log.Printf(
			"Listening for IP address changes on %s\n",
			strings.Join(w.interfaces, ", "),
		)
	} else {
		w.log.Println(
			"Listening for IP address changes on all interfaces",
		)
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
				if err := w.handleDelAddr(msg); err != nil {
					w.log.Println(err)
				}
			case unix.RTM_NEWADDR:
				runHooks := true
				if err := w.handleNewAddr(msg, runHooks); err != nil {
					w.log.Println(err)
				}
			}
		}
	}
}
