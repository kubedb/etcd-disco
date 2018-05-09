package cmds

import (
	"github.com/appscode/go/term"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdStop() *cobra.Command {
	opts := options.NewEtcdClusterConfig()
	etcdConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "stop",
		Short:             "Stop etcd cluster",
		Example:           "lector cluster stop <name>",
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
			server, err := etcd.NewServer(opts.ServerConfig, etcdConf)
			if err != nil {
				term.Fatalln(err)
			}
			if err := server.Seed(nil); err != nil {
				term.Fatalln(err)
			}

			select {}
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(etcdConf.Cf.FlagSet)

	return cmd
}
