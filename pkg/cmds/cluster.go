package cmds

import (
	"flag"
	"log"
	"strings"

	"github.com/appscode/go/term"
	"github.com/appscode/kutil/tools/analytics"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/etcd-manager/lector/pkg/operator"
	"github.com/etcd-manager/lector/pkg/providers/snapshot"
	_ "github.com/etcd-manager/lector/pkg/providers/snapshot/file"
	"github.com/jpillora/go-ogle-analytics"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	gaTrackingCode = "UA-62096468-20"
)

func NewCmdCluster() *cobra.Command {
	opts := options.NewEtcdClusterConfig()
	etcdConf := etcdmain.NewConfig()
	var (
		enableAnalytics = true
	)
	cmd := &cobra.Command{
		Use:               "etcd",
		Short:             "Create etcd cluster",
		Example:           "lector cluster create <name>",
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if enableAnalytics && gaTrackingCode != "" {
				if client, err := ga.NewClient(gaTrackingCode); err == nil {
					client.ClientID(analytics.ClientID())
					parts := strings.Split(c.CommandPath(), " ")
					client.Send(ga.NewEvent(parts[0], strings.Join(parts[1:], "/")).Label(""))
				}
			}
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
		//ClusterSize:             opts.ClusterSize,
		CurrentMemberAddress: opts.SelfAddrss,
	}

	operator.New(conf).Run()

}
