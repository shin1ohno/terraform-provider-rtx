package fwhelpers

import (
	"reflect"
	"testing"
)

// The contract mirrors reorderAddressesToMatchPlan in
// internal/provider/resources/ipv6_interface/model.go: normalize ordering, keep
// router-only values as drift, never fabricate plan-only values.

func TestReorderIntsToMatchPlan_AlreadyInPlanOrderIsUnchanged(t *testing.T) {
	got := ReorderIntsToMatchPlan([]int{6, 31, 36}, []int{6, 31, 36})
	if want := []int{6, 31, 36}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_RouterOrderIsNormalizedToPlanOrder(t *testing.T) {
	// The RTX prints the dynamic list in its own order; the plan order wins.
	got := ReorderIntsToMatchPlan([]int{36, 6, 31}, []int{6, 31, 36})
	if want := []int{6, 31, 36}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_FullyReversed(t *testing.T) {
	got := ReorderIntsToMatchPlan([]int{51, 46, 41, 36, 31, 6}, []int{6, 31, 36, 41, 46, 51})
	if want := []int{6, 31, 36, 41, 46, 51}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_RouterOnlyValueIsKeptLast(t *testing.T) {
	// A sequence added on the device is real drift and must stay visible, or the
	// whole point of reading from the router is lost.
	got := ReorderIntsToMatchPlan([]int{99, 31, 6}, []int{6, 31})
	if want := []int{6, 31, 99}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_PlanOnlyValueIsNotFabricated(t *testing.T) {
	// A sequence removed on the device must surface as a diff, not be papered over.
	got := ReorderIntsToMatchPlan([]int{6}, []int{6, 31, 36})
	if want := []int{6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_DuplicatesEachConsumedOnce(t *testing.T) {
	got := ReorderIntsToMatchPlan([]int{31, 6, 6}, []int{6, 6, 31})
	if want := []int{6, 6, 31}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_MoreDuplicatesOnRouterThanInPlan(t *testing.T) {
	got := ReorderIntsToMatchPlan([]int{6, 6, 6}, []int{6})
	if want := []int{6, 6, 6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_EmptyInputsAreNoOps(t *testing.T) {
	if got := ReorderIntsToMatchPlan(nil, []int{6}); got != nil {
		t.Fatalf("nil actual: got %v, want nil", got)
	}
	// No desired order to impose — e.g. a fresh import — so the router order stands.
	got := ReorderIntsToMatchPlan([]int{36, 6}, nil)
	if want := []int{36, 6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nil desired: got %v, want %v", got, want)
	}
}

func TestReorderIntsToMatchPlan_DisjointSetsKeepRouterOrder(t *testing.T) {
	got := ReorderIntsToMatchPlan([]int{41, 46}, []int{6, 31})
	if want := []int{41, 46}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
