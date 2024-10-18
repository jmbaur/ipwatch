package ipwatch

import (
	"net/netip"
	"testing"
)

func TestPassesFilter(t *testing.T) {
	tt := []struct {
		name        string
		inputAddr   netip.Addr
		inputFilter string
		expect      bool
	}{
		{
			name:        "::1 IsLoopback",
			inputAddr:   netip.MustParseAddr("::1"),
			inputFilter: "IsLoopback",
			expect:      true,
		},
		{
			name:        "::1 !IsLoopback",
			inputAddr:   netip.MustParseAddr("::1"),
			inputFilter: "!IsLoopback",
			expect:      false,
		},
	}

	for _, tc := range tt {
		if got := passesFilters(tc.inputAddr, tc.inputFilter); got != tc.expect {
			t.Fatalf("%s: got: %v, wanted: %v\n", tc.name, got, tc.expect)
		}
	}
}
