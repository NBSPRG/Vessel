package cgroups

import (
	"strconv"
	"testing"
)

// allTiers lists every defined tier for table-driven tests.
var allTiers = []Tier{TierMicro, TierSmall, TierMedium, TierLarge, TierXLarge}

// TestTierSpecsNonZero ensures every tier has positive resource values.
func TestTierSpecsNonZero(t *testing.T) {
	for _, tier := range allTiers {
		spec, ok := GetTierSpec(tier)
		if !ok {
			t.Errorf("tier %d (%s): not found in tierSpecs", tier, TierName(tier))
			continue
		}
		if spec.MemoryMB <= 0 {
			t.Errorf("tier %s: MemoryMB must be > 0, got %d", TierName(tier), spec.MemoryMB)
		}
		if spec.SwapMB <= 0 {
			t.Errorf("tier %s: SwapMB must be > 0, got %d", TierName(tier), spec.SwapMB)
		}
		if spec.CPUs <= 0 {
			t.Errorf("tier %s: CPUs must be > 0, got %f", TierName(tier), spec.CPUs)
		}
		if spec.PIDs <= 0 {
			t.Errorf("tier %s: PIDs must be > 0, got %d", TierName(tier), spec.PIDs)
		}
	}
}

// TestTierSpecsOrdering verifies that higher tiers have more resources.
func TestTierSpecsOrdering(t *testing.T) {
	ordered := []Tier{TierMicro, TierSmall, TierMedium, TierLarge, TierXLarge}
	for i := 1; i < len(ordered); i++ {
		prev, _ := GetTierSpec(ordered[i-1])
		curr, _ := GetTierSpec(ordered[i])
		if curr.MemoryMB <= prev.MemoryMB {
			t.Errorf("tier %s MemoryMB (%d) should be > tier %s MemoryMB (%d)",
				TierName(ordered[i]), curr.MemoryMB,
				TierName(ordered[i-1]), prev.MemoryMB)
		}
		if curr.PIDs <= prev.PIDs {
			t.Errorf("tier %s PIDs (%d) should be > tier %s PIDs (%d)",
				TierName(ordered[i]), curr.PIDs,
				TierName(ordered[i-1]), prev.PIDs)
		}
	}
}

// TestApplyTierSetsMem verifies that ApplyTier writes the correct memory bytes
// into the CGroups struct (memory in bytes = MemoryMB * MB).
func TestApplyTierSetsMem(t *testing.T) {
	for _, tier := range allTiers {
		spec, _ := GetTierSpec(tier)
		cg := NewCGroup().SetPath("test")
		cg.ApplyTier(tier)

		wantMem := strconv.Itoa(spec.MemoryMB * MB)
		gotMem := string(cg.mem)
		if gotMem != wantMem {
			t.Errorf("tier %s: mem = %q, want %q", TierName(tier), gotMem, wantMem)
		}
	}
}

// TestApplyTierSetsPIDs verifies that ApplyTier writes the correct PID limit.
func TestApplyTierSetsPIDs(t *testing.T) {
	for _, tier := range allTiers {
		spec, _ := GetTierSpec(tier)
		cg := NewCGroup().SetPath("test")
		cg.ApplyTier(tier)

		wantPIDs := strconv.Itoa(spec.PIDs)
		gotPIDs := string(cg.pids)
		if gotPIDs != wantPIDs {
			t.Errorf("tier %s: pids = %q, want %q", TierName(tier), gotPIDs, wantPIDs)
		}
	}
}

// TestApplyTierOverride verifies that an explicit SetMemorySwapLimit called
// after ApplyTier overrides the tier value — matching Kubernetes behaviour
// where per-pod resource requests override LimitRange defaults.
func TestApplyTierOverride(t *testing.T) {
	const customMemMB = 768
	const customSwapMB = 128

	cg := NewCGroup().SetPath("test")
	cg.ApplyTier(TierSmall) // 256 MB

	// Override with a custom value
	cg.SetMemorySwapLimit(customMemMB*MB, customSwapMB*MB)

	wantMem := strconv.Itoa(customMemMB * MB)
	gotMem := string(cg.mem)
	if gotMem != wantMem {
		t.Errorf("override: mem = %q, want %q", gotMem, wantMem)
	}
}

// TestParseTierValid checks that all tier name strings round-trip correctly.
func TestParseTierValid(t *testing.T) {
	cases := []struct {
		input string
		want  Tier
	}{
		{"micro", TierMicro},
		{"small", TierSmall},
		{"medium", TierMedium},
		{"large", TierLarge},
		{"xlarge", TierXLarge},
		{"MICRO", TierMicro},   // case-insensitive
		{"Medium", TierMedium}, // mixed case
	}
	for _, c := range cases {
		got, err := ParseTier(c.input)
		if err != nil {
			t.Errorf("ParseTier(%q) unexpected error: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseTier(%q) = %d, want %d", c.input, got, c.want)
		}
	}
}

// TestParseTierInvalid checks that unknown tier names return an error.
func TestParseTierInvalid(t *testing.T) {
	invalid := []string{"", "xxxx", "2xlarge", "tiny"}
	for _, name := range invalid {
		if _, err := ParseTier(name); err == nil {
			t.Errorf("ParseTier(%q) expected error, got nil", name)
		}
	}
}

// TestTierName verifies TierName returns non-empty strings for all tiers.
func TestTierName(t *testing.T) {
	for _, tier := range allTiers {
		if name := TierName(tier); name == "" || name == "unknown" {
			t.Errorf("TierName(%d) = %q, expected a valid name", tier, name)
		}
	}
}
