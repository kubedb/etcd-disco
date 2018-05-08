package cmds

import (
	"github.com/appscode/go/term"
	. "github.com/etcd-manager/lector/pkg"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
	"github.com/spf13/cobra"
)

func NewCmdJoin() *cobra.Command {
	opts := options.NewEtcdServerJoinConfig()
	etcdmainConf := etcdmain.NewConfig()
	cmd := &cobra.Command{
		Use:               "join",
		Short:             "Join a member to etcd cluster",
		Example:           "lector cluster join <name>",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.ValidateFlags(cmd, args); err != nil {
				term.Fatalln(err)
			}
			cfg := EtcdServerConfig(etcdmainConf)
			cfg.Name = opts.Name

			client, err := etcd.NewClient(opts.InitialUrls, cfg.ClientSC, true)
			if err != nil {
				term.Fatalln(err)
			}

			server := etcd.NewServer(cfg)
			if err := server.Join(client); err != nil {
				term.Fatalln(err)
			}
		},
	}
	opts.AddFlags(cmd.Flags())

	return cmd
}
