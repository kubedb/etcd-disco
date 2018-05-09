package cmds

import (
	"fmt"

	"github.com/appscode/go/term"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	opts := options.NewEtcdClusterConfig()
	etcdConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "create",
		Short:             "Create etcd cluster",
		Example:           "lector cluster create <name>",
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

			snap, err := server.SnapshotInfo()
			if err != nil {
				fmt.Println(err)
			}
			//snap = nil // TODO(check)::
			if err := server.Seed(snap); err != nil {
				term.Fatalln(err)
			}

			select {}
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(etcdConf.Cf.FlagSet)

	return cmd
}
