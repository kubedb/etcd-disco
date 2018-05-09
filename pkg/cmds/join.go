package cmds

import (
	"fmt"

	"github.com/appscode/go/term"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/spf13/cobra"
)

func NewCmdJoin() *cobra.Command {
	opts := options.NewEtcdServerJoinConfig()
	cmd := &cobra.Command{
		Use:               "join",
		Short:             "Join a member to etcd cluster",
		Example:           "lector cluster join <name>",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.ValidateFlags(cmd, args); err != nil {
				term.Fatalln(err)
			}
			cfg := opts.EtcdServerConfig()
			cfg.Name = opts.Name

			client, err := etcd.NewClient([]string{opts.ServerAddress}, cfg.ClientSC, true)
			if err != nil {
				fmt.Println(opts.ServerAddress)
				term.Fatalln(err, "***")
			}

			server := etcd.NewServer(cfg)
			if err := server.Join(client); err != nil {
				term.Fatalln(err)
			}
			select {}
		},
	}
	opts.AddFlags(cmd.Flags())

	return cmd
}
