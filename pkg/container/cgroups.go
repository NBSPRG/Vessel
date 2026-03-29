package container

import (
	"fmt"
	"github.com/0xc0d/vessel/pkg/cgroups"
	"path/filepath"
)

// LoadCGroups loads CGroups for container.
// If a resource tier is set it is applied first; any explicit per-field
// limits (memory, swap, cpu, pids) are then applied on top, overriding
// the tier defaults — mirroring Kubernetes LimitRange override semantics.
func (c *Container) LoadCGroups() error {
	cg := cgroups.NewCGroup()
	cg.SetPath(filepath.Join("vessel", c.Digest))

	// Apply tier baseline first (no-op when tier == -1)
	if c.tier >= 0 {
		cg.ApplyTier(cgroups.Tier(c.tier))
	}

	// Explicit limits override tier defaults (zero means "use tier value")
	if c.mem > 0 || c.swap > 0 {
		cg.SetMemorySwapLimit(c.mem, c.swap)
	}
	if c.cpus > 0 {
		cg.SetCPULimit(c.cpus)
	}
	if c.pids > 0 {
		cg.SetProcessLimit(c.pids)
	}

	err := cg.Load()
	if err != nil {
		return err
	}
	pids, err := cg.GetPids()
	c.Pids = pids
	return err
}

// RemoveCGroups removes CGroups file for container.
// It only function if the container is not running.
func (c *Container) removeCGroups() error {
	cg := &cgroups.CGroups{
		Path: filepath.Join("vessel", c.Digest),
	}
	return cg.Remove()
}

// SetMemorySwapLimit sets Container's memory and swap limitation in MegaByte.
func (c *Container) SetMemorySwapLimit(memory, swap int) *Container {
	c.mem = memory * MB
	c.swap = swap * MB
	return c
}

// SetCPULimit sets Container number of CPUs.
func (c *Container) SetCPULimit(cpus float64) *Container {
	c.cpus = cpus
	return c
}

// SetProcessLimit sets maximum simultaneous process for Container.
func (c *Container) SetProcessLimit(pids int) *Container {
	c.pids = pids
	return c
}

// GetPids returns slice of pid running inside Container.
//
// NOTE: First element [0], is the fork process.
func (c *Container) GetPids() ([]int, error) {
	cg := &cgroups.CGroups{
		Path: filepath.Join("vessel", c.Digest),
	}
	pids, err := cg.GetPids()
	return pids, err
}

// getPidsByDigest returns slice of pid running inside a Container.
// Container should be specified by its digest.
func getPidsByDigest(digest string) ([]int, error) {
	cg := &cgroups.CGroups{
		Path: filepath.Join("vessel", digest),
	}
	pids, err := cg.GetPids()
	return pids, err
}

// SetResourceTier configures a named resource tier (micro/small/medium/large/xlarge).
// The tier sets baseline CPU, memory, and PID limits; any explicit
// SetMemorySwapLimit / SetCPULimit / SetProcessLimit calls made afterward
// will override the tier values for that specific field.
func (c *Container) SetResourceTier(name string) error {
	tier, err := cgroups.ParseTier(name)
	if err != nil {
		return fmt.Errorf("invalid resource tier: %w", err)
	}
	c.tier = int(tier)
	return nil
}
