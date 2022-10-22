package ipwatch

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

var (
	ErrInvalidIpProtocol = errors.New("only one of -4 and -6 allowed")
	ErrNotImplemented    = errors.New("not implemented")
	ErrInvalidFilter     = errors.New("invalid filter")
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

func isValidFilter(filter string) bool {
	test := filter
	if strings.HasPrefix(filter, "!") {
		test = filter[1:]
	}
	_, ok := validFilters[test]
	return ok
}

func passesFilter(addr netip.Addr, filter string) bool {
	var not bool
	if strings.HasPrefix(filter, "!") {
		not = true
		filter = filter[1:]
	}

	var result bool
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
		return !result
	}

	return result
}

type WatchConfig struct {
	MaxRetries uint
	Interfaces []string
	Hooks      []string
	IPv4       bool
	IPv6       bool
	Filters    []string
}

type Watcher struct {
	interfaces []string
	hooks      []Hook
	filters    []string
	maxRetries uint
}

func NewWatcher(cfg WatchConfig) (*Watcher, error) {
	if cfg.IPv4 && cfg.IPv6 {
		return nil, ErrInvalidIpProtocol
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
		} else {
			filterMap[filter] = struct{}{}
		}
	}

	filters := []string{}
	for k := range filterMap {
		filters = append(filters, k)
	}

	return &Watcher{
		interfaces: cfg.Interfaces,
		hooks:      hooks,
		filters:    filters,
		maxRetries: cfg.MaxRetries,
	}, nil
}

func (w *Watcher) Watch() error {
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
			return fmt.Errorf("failed to receive messages: %w", err)
		}
	messages:
		for _, msg := range msgs {
			if msg.Header.Type != unix.RTM_NEWADDR {
				continue
			}

			if len(msg.Data) < unix.SizeofIfAddrmsg {
				continue
			}

			var (
				iface *net.Interface
				newIP *netip.Addr
			)

			// struct ifaddrmsg {
			//     unsigned char ifa_family;    /* Address type */
			//     unsigned char ifa_prefixlen; /* Prefixlength of address */
			//     unsigned char ifa_flags;     /* Address flags */
			//     unsigned char ifa_scope;     /* Address scope */
			//     unsigned int  ifa_index;     /* Interface index */
			// };
			iface, err = net.InterfaceByIndex(int(msg.Data[4]))
			if err != nil {
				continue
			}

			interested := len(w.interfaces) == 0
			for _, i := range w.interfaces {
				interested = i == iface.Name
			}
			if !interested {
				continue
			}

			ad, err := netlink.NewAttributeDecoder(msg.Data[unix.SizeofIfAddrmsg:])
			if err != nil {
				continue
			}

			for ad.Next() {
				if newIP != nil {
					break
				}
				switch ad.Type() {
				case unix.IFA_ADDRESS:
					{
						ip := ad.Bytes()
						addr, ok := netip.AddrFromSlice(ip)
						if !ok {
							continue messages
						}
						newIP = &addr
						for _, filter := range w.filters {
							if !passesFilter(*newIP, filter) {
								continue messages
							}
						}
					}
				}
			}

			if newIP == nil {
				continue messages
			}

			var wg sync.WaitGroup
			var l sync.Mutex

			for _, hook := range w.hooks {
				wg.Add(1)
				hook := hook
				go func() {
					for i := 1; i <= int(w.maxRetries); i++ {
						backoff := time.Duration(math.Pow(2, float64(i))) * time.Second

						if hookOutput, err := hook.Run(iface.Name, *newIP); err != nil {
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
		}
	}
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
