package cmds

import (
	"github.com/etcd-manager/lector/pkg/cmds"
	"github.com/spf13/cobra"
)

func newCmdCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "etcd",
		DisableAutoGenTag: true,
		Run:               func(cmd *cobra.Command, args []string) {},
	}

	cmd.AddCommand(cmds.NewCmdCreate())
	cmd.AddCommand(cmds.NewCmdJoin())
	cmd.AddCommand(cmds.NewCmdSnapshot())
	cmd.AddCommand(cmds.NewCmdStop())

	return cmd
}
