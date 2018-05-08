package options

import (
	"github.com/appscode/go/flags"
	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type EtcdServerJoinConfig struct {
	Name string

	Config      string
	InitialUrls []string
	ClientUrl   string
	PeerUrl     string
	MetricUrls  []string

	ClusterState string
}

func NewEtcdServerJoinConfig() *EtcdServerJoinConfig {
	return &EtcdServerJoinConfig{
		InitialUrls: []string{},
		ClientUrl:   "",
		PeerUrl:     "",
		MetricUrls:  []string{},

		ClusterState: embed.ClusterStateFlagExisting,
	}
}

func (s *EtcdServerJoinConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&s.InitialUrls, "initial-urls", "i", s.InitialUrls, "list of initial URLs to join current member. e.g. http://localhost:2379")
	fs.StringVarP(&s.ClientUrl, "client-urls", "c", s.ClientUrl, "list of URLs to listen on for client traffic. e.g. http://localhost:2379")
	fs.StringVarP(&s.PeerUrl, "peer-urls", "p", s.PeerUrl, "list of URLs to listen on for peer traffic. e.g. http://localhost:2379")
	fs.StringArrayVarP(&s.MetricUrls, "metric-urls", "m", s.MetricUrls, "list of URLs to listen on for metric traffic. e.g. http://localhost:2379")
	//fs.StringVar(&s.Config, "config", "", "member config file")
}

func (s *EtcdServerJoinConfig) ValidateFlags(cmd *cobra.Command, args []string) error {
	ensureFlags := []string{"initial-urls", "client-url", "peer-url", "metric-urls"}
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
