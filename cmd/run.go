package cmd

import (
	"github.com/0xc0d/vessel/internal"
	"github.com/spf13/cobra"
)

// NewRunCommand implements and returns the run command.
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
		Short:                 "Run a command inside a new Container.",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  internal.Run,
	}

	flags := cmd.Flags()
	flags.StringP("host", "", "", "Container Hostname")
	flags.IntP("memory", "m", 0, "Limit memory access in MB (overrides --tier)")
	flags.IntP("swap", "s", 0, "Limit swap access in MB (overrides --tier)")
	flags.Float64P("cpus", "c", 0, "Limit CPUs (overrides --tier)")
	flags.IntP("pids", "p", 0, "Limit number of processes (overrides --tier)")
	flags.BoolP("detach", "d", false, "run command in the background")
	flags.StringP("tier", "t", "", "Resource tier: micro|small|medium|large|xlarge")

	return cmd
}
