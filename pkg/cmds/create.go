package cmds

import (
	"github.com/appscode/go/term"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	opts := options.NewEtcdServerCreateConfig()
	cmd := &cobra.Command{
		Use:               "create",
		Short:             "Create etcd cluster",
		Example:           "lector cluster create <name>",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.ValidateFlags(cmd, args); err != nil {
				term.Fatalln(err)
			}
			cfg := opts.EtcdServerConfig()
			cfg.Name = opts.Name

			server := etcd.NewServer(cfg)
			if err := server.Seed(nil); err != nil {
				term.Fatalln(err)
			}

			select {}
		},
	}
	//opts.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(etcdmain.NewConfig().Cf.FlagSet)

	return cmd
}
