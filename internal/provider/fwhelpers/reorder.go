package fwhelpers

// ReorderIntsToMatchPlan reorders a read-back int list so that values also present
// in desired appear in desired's order.
//
// Terraform treats a ListAttribute as ordered, so a read-back that merely permutes
// the configured values is rejected as "Provider produced inconsistent result after
// apply". The RTX does not echo a `secure filter ... dynamic <n...>` list in the
// order it was written, so the normalization has to happen provider-side.
//
// desired is the plan on Create/Update and the PRIOR STATE on Read. Passing prior
// state on Read matters: without it a refresh rewrites state in router order and
// every subsequent plan shows a spurious reordering diff.
//
// This only normalizes ordering:
//
//   - a value the router returned that desired does not mention is KEPT, appended
//     after the matched ones — it is real drift and has to stay visible;
//   - a value in desired the router did not return is NOT fabricated, so a sequence
//     removed on the device still surfaces as a diff;
//   - repeated identical values are consumed one-for-one.
//
// Same bug class and remedy as reorderAddressesToMatchPlan in
// internal/provider/resources/ipv6_interface/model.go and
// reorderServerSelectToMatchPlan in internal/provider/resources/dns_server/model.go.
// Those two key on a struct because their elements are nested blocks; here the
// element is a bare int, so the value is its own key. It lives here rather than in
// either resource because the IPv4 and IPv6 access lists need the identical
// operation on the identical type.
func ReorderIntsToMatchPlan(actual, desired []int) []int {
	if len(desired) == 0 || len(actual) == 0 {
		return actual
	}

	// One index list per value so repeated identical values are consumed once each.
	byValue := make(map[int][]int, len(actual))
	for i, v := range actual {
		byValue[v] = append(byValue[v], i)
	}

	reordered := make([]int, 0, len(actual))
	matched := make([]bool, len(actual))

	for _, want := range desired {
		indices := byValue[want]
		if len(indices) == 0 {
			continue
		}
		idx := indices[0]
		byValue[want] = indices[1:]
		matched[idx] = true
		reordered = append(reordered, actual[idx])
	}

	for i, v := range actual {
		if !matched[i] {
			reordered = append(reordered, v)
		}
	}

	return reordered
}
