package cmds

import (
	"flag"
	"log"

	"github.com/appscode/etcd-disco/pkg/cmds/options"
	"github.com/appscode/etcd-disco/pkg/etcdmain"
	"github.com/appscode/etcd-disco/pkg/operator"
	"github.com/appscode/etcd-disco/pkg/providers/snapshot"
	_ "github.com/appscode/etcd-disco/pkg/providers/snapshot/file"
	"github.com/appscode/go/term"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewCmdCluster() *cobra.Command {
	opts := options.NewEtcdClusterConfig()
	etcdConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "etcd-disco",
		Short:             "Create etcd cluster",
		Example:           "etcd-disco cluster create <name>",
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.ValidateFlags(cmd, args); err != nil {
				term.Fatalln(err)
			}
			if err := etcdConf.ConfigFromCmdLine(); err != nil {
				term.Fatalln(err)
			}
			Start(opts, etcdConf)
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	opts.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(etcdConf.Cf.FlagSet)
	flag.CommandLine.Parse([]string{})
	return cmd
}

func Start(opts *options.EtcdClusterConfig, etcdConf *etcdmain.Config) {
	if etcdConf.Ec.Dir == "" {
		etcdConf.Ec.Dir = "/tmp/etcd/" + etcdConf.Ec.Name
	}

	conf := operator.Config{
		Snapshot: snapshot.Config{
			Interval: opts.CheckInterval,
			TTL:      opts.SnapshotInterval,
			Provider: "file",
		},
		Etcd:                    etcdConf,
		UnhealthyMemberTTL:      opts.UnhealthyMemberTTL,
		InitialMembersAddresses: opts.ServerAddress,
		CurrentMemberAddress:    opts.SelfAddrss,
	}

	operator.New(conf).Run()

}
