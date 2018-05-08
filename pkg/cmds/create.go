package cmds

import (
	"github.com/appscode/go/term"
	. "github.com/etcd-manager/lector/pkg"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	opts := options.NewEtcdServerCreateConfig()
	etcdmainConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "create",
		Short:             "Create etcd cluster",
		Example:           "lector cluster create <name>",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.ValidateFlags(cmd, args); err != nil {
				term.Fatalln(err)
			}
			cfg := EtcdServerConfig(etcdmainConf)
			cfg.Name = opts.Name

			server := etcd.NewServer(cfg)
			if err := server.Seed(nil); err != nil {
				term.Fatalln(err)
			}
		},
	}
	opts.AddFlags(cmd.Flags())

	return cmd
}
