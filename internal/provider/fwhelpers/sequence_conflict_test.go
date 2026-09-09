package fwhelpers

import (
	"reflect"
	"testing"
)

func TestCheckSequenceContentConflicts(t *testing.T) {
	// The wan-outbound-ipv6-dynamic rows as the RTX1210 rendered them on 2026-09-08.
	router := map[int]string{
		1: "* * ftp", 6: "dhcp-prefix@lan2::/64 * domain", 11: "* * www",
		16: "* * smtp", 21: "* * pop3", 26: "* * submission",
		31: "dhcp-prefix@lan2::/64 * tcp", 36: "dhcp-prefix@lan2::/64 * udp",
		41: "ra-prefix@lan2::/64 * domain", 46: "ra-prefix@lan2::/64 * tcp", 51: "ra-prefix@lan2::/64 * udp",
	}
	same := func() map[int]string {
		m := make(map[int]string, len(router))
		for k, v := range router {
			m[k] = v
		}
		return m
	}
	withRow := func(seq int, content string) map[int]string {
		m := same()
		m[seq] = content
		return m
	}

	tests := []struct {
		name         string
		planned      map[int]string
		existing     map[int]string
		currentState []int
		want         []int
	}{
		{
			// Create after `terraform state rm`: nothing is owned, everything is on the
			// router, all of it identical. The bare-sequence check reports all 11.
			name:     "create over identical rows is silent",
			planned:  same(),
			existing: router,
			want:     nil,
		},
		{
			// The 2026-09-08 incident: state knew 8 rows, config and router had 11.
			name:         "update with a short state is silent when the extra rows match",
			planned:      same(),
			existing:     router,
			currentState: []int{1, 6, 11, 16, 21, 26, 31, 36},
			want:         nil,
		},
		{
			name:         "unowned row with different content is a conflict",
			planned:      same(),
			existing:     withRow(41, "* * domain"),
			currentState: []int{1, 6, 11, 16, 21, 26, 31, 36},
			want:         []int{41},
		},
		{
			name:         "owned row with different content is an update, not a conflict",
			planned:      same(),
			existing:     withRow(6, "* * domain"),
			currentState: []int{1, 6, 11, 16, 21, 26, 31, 36},
			want:         nil,
		},
		{
			name:     "planned rows the router does not have are free",
			planned:  map[int]string{100: "* * ftp", 105: "* * www"},
			existing: router,
			want:     nil,
		},
		{
			name:     "no planned rows",
			planned:  map[int]string{},
			existing: router,
			want:     nil,
		},
		{
			name:     "conflicts come back sorted",
			planned:  map[int]string{51: "x", 1: "x", 26: "x", 100: "x"},
			existing: router,
			want:     []int{1, 26, 51},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CheckSequenceContentConflicts(tc.planned, tc.existing, tc.currentState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("conflicts = %v, want %v", got, tc.want)
			}
		})
	}
}

// The bare-sequence check is what the content-aware one replaces in the dynamic
// filter resources; pin the behaviour that made the replacement necessary so the
// difference between the two stays documented by a test.
func TestCheckSequenceConflicts_FlagsOwnRowsWhenStateIsShort(t *testing.T) {
	planned := []int{1, 6, 11, 16, 21, 26, 31, 36, 41, 46, 51}
	got := CheckSequenceConflicts(planned, planned, planned[:8])
	if want := []int{41, 46, 51}; !reflect.DeepEqual(got, want) {
		t.Fatalf("conflicts = %v, want %v", got, want)
	}
	if got := CheckSequenceConflicts(planned, planned, nil); !reflect.DeepEqual(got, planned) {
		t.Fatalf("create-path conflicts = %v, want every sequence", got)
	}
}
