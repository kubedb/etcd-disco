package cmds

import (
	"github.com/etcd-manager/lector/pkg/cmds"
	"github.com/spf13/cobra"
)

func newCmdCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cluster",
		DisableAutoGenTag: true,
		Run:               func(cmd *cobra.Command, args []string) {},
	}

	cmd.AddCommand(cmds.NewCmdCreate())
	cmd.AddCommand(cmds.NewCmdJoin())

	return cmd
}