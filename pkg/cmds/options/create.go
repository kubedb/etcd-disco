package options

import (
	"github.com/appscode/go/flags"
	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type EtcdServerCreateConfig struct {
	Name string

	ClusterState string
}

func NewEtcdServerCreateConfig() *EtcdServerCreateConfig {
	return &EtcdServerCreateConfig{
		ClusterState: embed.ClusterStateFlagNew,
	}
}

func (s *EtcdServerCreateConfig) AddFlags(fs *pflag.FlagSet) {
	//fs.StringVar(&s.Config, "config", "", "member config file")
}

func (s *EtcdServerCreateConfig) ValidateFlags(cmd *cobra.Command, args []string) error {
	ensureFlags := []string{}
	//ensureFlags := []string{"config"}
	flags.EnsureRequiredFlags(cmd, ensureFlags...)

	if len(args) == 0 {
		errors.New("missing member name")
	}
	if len(args) > 1 {
		errors.New("multiple member name provided")
	}
	s.Name = args[0]
	return nil
}
