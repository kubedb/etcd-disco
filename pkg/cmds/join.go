package cmds

import (
	//"fmt"
	"github.com/appscode/go/term"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	//"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/spf13/cobra"
	//"github.com/Masterminds/glide/cfg"
	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/etcd-manager/lector/pkg/etcdmain"
)

func NewCmdJoin() *cobra.Command {
	opts := options.NewEtcdClusterConfig()
	etcdConf := etcdmain.NewConfig()
	var ServerAddress string
	cmd := &cobra.Command{
		Use:               "join",
		Short:             "Join a member to etcd cluster",
		Example:           "lector cluster join <name>",
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

			client, err := etcd.NewClient([]string{ServerAddress}, etcd.SecurityConfig{
				CAFile:        etcdConf.Ec.ClientTLSInfo.CAFile,
				CertFile:      etcdConf.Ec.ClientTLSInfo.CertFile,
				KeyFile:       etcdConf.Ec.ClientTLSInfo.KeyFile,
				CertAuth:      etcdConf.Ec.ClientTLSInfo.ClientCertAuth,
				TrustedCAFile: etcdConf.Ec.ClientTLSInfo.TrustedCAFile,
				AutoTLS:       etcdConf.Ec.ClientAutoTLS,
			}, true)
			if err != nil {
				term.Fatalln(err)
			}
			server, err := etcd.NewServer(opts.ServerConfig, etcdConf)
			if err != nil {
				term.Fatalln(err)
			}
			if err := server.Join(client); err != nil {
				term.Fatalln(err)
			}

			select {}
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(etcdConf.Cf.FlagSet)
	cmd.Flags().StringVar(&ServerAddress, "server-address", "", "List of URLs to listen on for peer traffic.")

	return cmd
}
