package cmd

import (
	"github.com/0xc0d/vessel/internal"
	"github.com/0xc0d/vessel/pkg/container"
	"github.com/spf13/cobra"
)

// NewForkCommand implements and returns fork command.
// fork command is called by reexec to apply namespaces.
//
// It is a hidden command and requires root path and
// container id to run.
func NewForkCommand() *cobra.Command {
	ctr := container.NewContainer()
	var detach bool
	cmd := &cobra.Command{
		Use:          "fork",
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ctr.LoadConfig(); err != nil {
				return err
			}
			return internal.Fork(ctr, args, detach)
		},
	}

	var (
		mem  int
		swap int
		cpu  float64
		pids int
		tier string
	)

	flags := cmd.Flags()
	flags.StringVar(&ctr.Digest, "container", "", "")
	flags.StringVar(&ctr.RootFS, "root", "", "")
	flags.StringVar(&ctr.Config.Hostname, "host", "", "")
	flags.BoolVar(&detach, "detach", false, "")
	flags.IntVar(&mem, "memory", 0, "")
	flags.IntVar(&swap, "swap", 0, "")
	flags.Float64Var(&cpu, "cpus", 0, "")
	flags.IntVar(&pids, "pids", 0, "")
	flags.StringVar(&tier, "tier", "", "")

	// Wrap RunE to apply resource settings after cobra has parsed the flags.
	baseRunE := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if tier != "" {
			if err := ctr.SetResourceTier(tier); err != nil {
				return err
			}
		}
		// Explicit limits override tier (zero = defer to tier)
		if mem > 0 || swap > 0 {
			ctr.SetMemorySwapLimit(mem, swap)
		}
		if cpu > 0 {
			ctr.SetCPULimit(cpu)
		}
		if pids > 0 {
			ctr.SetProcessLimit(pids)
		}
		return baseRunE(c, args)
	}

	if err := cmd.MarkFlagRequired("root"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("container"); err != nil {
		panic(err)
	}
	return cmd
}
