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
	"syscall"
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

// WatcherConfig controls how the watcher behaves.
type WatcherConfig struct {
	Debug bool
}

// WatchConfig sets filters and hooks for the watcher.
type WatchConfig struct {
	Filters         []string
	Hooks           []string
	Interfaces      []string
	MaxRetries      uint
	HookEnvironment []string
}

// Watcher can be used to watch changes to IP addresses, optionally filtering
// on interface and/or IP address types.
type Watcher struct {
	log     *log.Logger
	l       sync.Mutex
	ipCache map[int]map[netip.Addr]struct{}
}

// NewWatcher parses a WatchConfig and returns a new Watcher.
func NewWatcher(cfg WatcherConfig) (*Watcher, error) {
	var logger *log.Logger
	if cfg.Debug {
		logger = log.Default()
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	return &Watcher{
		log:     logger,
		ipCache: map[int]map[netip.Addr]struct{}{},
	}, nil
}

func getIfInfomsg(ifinfomsg []byte) (*unix.IfInfomsg, error) {
	// struct ifinfomsg {
	//   unsigned char   ifi_family;
	//   unsigned char   __ifi_pad;
	//   unsigned short  ifi_type;   // ARPHRD_*
	//   int             ifi_index;  // Link index
	//   unsigned        ifi_flags;  // IFF_* flags
	//   unsigned        ifi_change; // IFF_* change mask
	// };
	return (*unix.IfInfomsg)(unsafe.Pointer(&ifinfomsg[0:unix.SizeofIfInfomsg][0])), nil
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

func (w *Watcher) handleNewAddr(msg netlink.Message, filters []string, hooks []Hook, maxRetries uint, startup bool) error {
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

				if cachedIface, ok := w.ipCache[idx]; ok {
					if _, ok := cachedIface[addr]; ok {
						w.log.Println("New addr was found in cache, skipping hooks")
						return nil
					}
				}

				newIP = addr
			}
		}
	}

	if !newIP.IsValid() {
		w.log.Println("No address found, skipping hooks")
		return nil
	}

	if !passesFilter(newIP, filters...) {
		w.log.Println("Address does not pass filters, skipping hooks")
		return nil
	}

	w.log.Println("Caching new address", newIP)
	w.cacheAddr(idx, newIP)

	if !fresh && startup {
		w.log.Println("IP address is not new and the program did not just startup, skipping hooks")
		return nil
	} else if fresh && startup {
		w.log.Println("Fresh IP address and starting up, running hooks")
	}

	var (
		wg sync.WaitGroup
		l  sync.Mutex
	)

	for _, hook := range hooks {
		wg.Add(1)
		hook := hook
		if maxRetries < 1 {
			maxRetries = 1
		}
		go func() {
			for i := 1; i <= int(maxRetries); i++ {
				backoff := time.Duration(math.Pow(2, float64(i))) * time.Second

				hookOutput, err := hook.Run(ifaddrmsg.Index, newIP)
				if err != nil {
					outputs := []string{}
					outputs = append(outputs, fmt.Sprintf("Hook '%s' failed", hook.Name()))
					if len(hookOutput) > 0 {
						outputs = append(outputs, hookOutput)
					}
					outputs = append(outputs, err.Error())
					if i < int(maxRetries) {
						outputs = append(outputs, fmt.Sprintf("Retrying in %s", backoff))
					} else {
						outputs = append(outputs, "Max attempts reached")
					}

					l.Lock()
					printOutput(outputs...)
					l.Unlock()

					time.Sleep(backoff)
					continue
				}

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

func (w *Watcher) generateCache(interfaces []string, filters []string, hooks []Hook, maxRetries uint) error {
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
				if len(interfaces) > 0 {
					for _, iface := range interfaces {
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
			startup := true
			if err := w.handleNewAddr(msg, filters, hooks, maxRetries, startup); err != nil {
				return err
			}
		}
	}

	return nil
}

// Watch watches for IP address changes performs hook actions on new IP
// addresses. This function blocks.
func (w *Watcher) Watch(cfg WatchConfig) error {
	if len(cfg.Hooks) == 0 {
		cfg.Hooks = append(cfg.Hooks, "internal:echo")
	}
	hooks := []Hook{}
	for _, hookName := range cfg.Hooks {
		hook, err := NewHook(hookName, cfg.HookEnvironment)
		if err != nil {
			return fmt.Errorf("%w: %s", err, hookName)
		}
		hooks = append(hooks, hook)
	}

	filterMap := map[string]struct{}{}
	for _, filter := range cfg.Filters {
		if !isValidFilter(filter) {
			return fmt.Errorf("%w: %s", ErrInvalidFilter, filter)
		}
		filterMap[filter] = struct{}{}
	}

	filters := []string{}
	for k := range filterMap {
		filters = append(filters, k)
	}

	if err := w.generateCache(cfg.Interfaces, filters, hooks, cfg.MaxRetries); err != nil {
		return err
	}

	if len(cfg.Interfaces) > 0 {
		w.log.Printf(
			"Listening for IP address changes on %s\n",
			strings.Join(cfg.Interfaces, ", "),
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
				startup := false
				if err := w.handleNewAddr(msg, filters, hooks, cfg.MaxRetries, startup); err != nil {
					w.log.Println(err)
				}
			}
		}
	}
}
