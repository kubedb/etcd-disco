package cmds

import (
	"github.com/appscode/go/term"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdSnapshot() *cobra.Command {
	opts := options.NewEtcdClusterConfig()
	etcdConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "snapshot",
		Short:             "Store etcd cluster snapshot",
		Example:           "lector cluster snapshot <name>",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.ValidateFlags(cmd, args); err != nil {
				term.Fatalln(err)
			}
			if err := etcdConf.ConfigFromCmdLine(); err != nil {
				term.Fatalln(err)
			}
			if etcdConf.Ec.Dir == "" {
				etcdConf.Ec.Dir = "/tmp/etcd/" + etcdConf.Ec.Name
			}

			etcdConf.Ec.InitialCluster = etcdConf.Ec.InitialClusterFromName(etcdConf.Ec.Name)
			server := etcd.NewServer(opts.ServerConfig, etcdConf)

			if err := server.Snapshot(); err != nil {
				term.Fatalln(err)
			}

			select {}
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(etcdConf.Cf.FlagSet)

	return cmd
}
