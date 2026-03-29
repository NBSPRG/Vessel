package cgroups

import (
	"fmt"
	"strings"
)

// MB is 1 mebibyte in bytes, used for tier memory conversions.
const MB = 1 << 20

// Tier represents a predefined resource allocation level, analogous to
// Kubernetes LimitRange / ResourceQuota tiers (micro → xlarge).
// Each tier provides a balanced CPU / memory / process budget suitable
// for different container workload classes.
type Tier int

const (
	TierMicro  Tier = iota // 128 MB RAM, 0.10 CPU, 50  PIDs — minimal scripts/utilities
	TierSmall              // 256 MB RAM, 0.25 CPU, 100 PIDs — lightweight services
	TierMedium             // 512 MB RAM, 0.50 CPU, 200 PIDs — standard workloads
	TierLarge              // 1024MB RAM, 1.00 CPU, 500 PIDs — data-processing jobs
	TierXLarge             // 2048MB RAM, 2.00 CPU, 1000 PIDs — memory-intensive workloads
)

// TierSpec holds the resource limits for a given tier.
// MemoryMB and SwapMB are in mebibytes; CPUs is fractional cores;
// PIDs is the maximum simultaneous process count.
type TierSpec struct {
	MemoryMB int
	SwapMB   int
	CPUs     float64
	PIDs     int
}

// tierSpecs is the authoritative map from Tier → resource limits.
// Swap is set to 50% of memory to limit swap pressure without
// disabling it entirely — mirrors Kubernetes memory request/limit patterns.
var tierSpecs = map[Tier]TierSpec{
	TierMicro:  {MemoryMB: 128, SwapMB: 64, CPUs: 0.10, PIDs: 50},
	TierSmall:  {MemoryMB: 256, SwapMB: 128, CPUs: 0.25, PIDs: 100},
	TierMedium: {MemoryMB: 512, SwapMB: 256, CPUs: 0.50, PIDs: 200},
	TierLarge:  {MemoryMB: 1024, SwapMB: 512, CPUs: 1.00, PIDs: 500},
	TierXLarge: {MemoryMB: 2048, SwapMB: 1024, CPUs: 2.00, PIDs: 1000},
}

// TierName returns a human-readable name for a Tier.
func TierName(t Tier) string {
	names := map[Tier]string{
		TierMicro:  "micro",
		TierSmall:  "small",
		TierMedium: "medium",
		TierLarge:  "large",
		TierXLarge: "xlarge",
	}
	if n, ok := names[t]; ok {
		return n
	}
	return "unknown"
}

// ParseTier converts a string name (case-insensitive) to a Tier value.
// Returns an error for unrecognised names.
func ParseTier(name string) (Tier, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "micro":
		return TierMicro, nil
	case "small":
		return TierSmall, nil
	case "medium":
		return TierMedium, nil
	case "large":
		return TierLarge, nil
	case "xlarge":
		return TierXLarge, nil
	default:
		return TierMicro, fmt.Errorf("unknown tier %q: valid values are micro, small, medium, large, xlarge", name)
	}
}

// GetTierSpec returns the TierSpec for a given Tier.
func GetTierSpec(t Tier) (TierSpec, bool) {
	spec, ok := tierSpecs[t]
	return spec, ok
}

// ApplyTier configures the CGroups instance with the limits defined for
// tier t.  Any subsequent call to SetMemorySwapLimit, SetCPULimit, or
// SetProcessLimit will override the tier values, enabling per-container
// fine-tuning on top of a baseline tier — mirroring how Kubernetes allows
// resource requests to override default LimitRange entries.
func (cg *CGroups) ApplyTier(t Tier) *CGroups {
	spec, ok := tierSpecs[t]
	if !ok {
		return cg
	}
	cg.SetMemorySwapLimit(spec.MemoryMB*MB, spec.SwapMB*MB)
	cg.SetCPULimit(spec.CPUs)
	cg.SetProcessLimit(spec.PIDs)
	return cg
}
