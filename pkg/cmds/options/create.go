package options

import (
	//	"net/url"
	"github.com/appscode/go/flags"
	"github.com/coreos/etcd/embed"
	//	"github.com/coreos/etcd/pkg/types"
	"time"

	"github.com/etcd-manager/lector/pkg/etcd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type EtcdServerCreateConfig struct {
	Name string

	Ec embed.Config

	NodeAddress string

	ClusterState string
}

func NewEtcdServerCreateConfig() *EtcdServerCreateConfig {
	return &EtcdServerCreateConfig{
		ClusterState: embed.ClusterStateFlagNew,
	}
}

func (cfg *EtcdServerCreateConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cfg.Ec.Dir, "data-dir", cfg.Ec.Dir, "Path to the data directory.")
	fs.StringVar(&cfg.NodeAddress, "node-address", "", "List of URLs to listen on for peer traffic.")
	//fs.StringArrayVar(&cfg.ClientUrls, "listen-client-urls", cfg.ClientUrls, "List of URLs to listen on for client traffic.")
	//fs.StringArrayVar(&cfg.MetricUrls, "listen-metric-urls", cfg.MetricUrls, "List of URLs to listen on for metrics.")
	//fs.StringVar(&cfg.Ec.ListenMetricsUrlsJSON, "listen-metrics-urls", "", "List of URLs to listen on for metrics.")
	fs.Int64Var(&cfg.Ec.QuotaBackendBytes, "quota-backend-bytes", 2147483648, "Raise alarms when backend size exceeds the given quota. 0 means use the default quota.")
	fs.StringVar(&cfg.ClusterState, "initial-cluster-state", embed.ClusterStateFlagNew, "Initial cluster state ('new' or 'existing').")

	fs.StringVar(&cfg.Ec.ClientTLSInfo.CAFile, "ca-file", "", "DEPRECATED: Path to the client server TLS CA file.")
	fs.StringVar(&cfg.Ec.ClientTLSInfo.CertFile, "cert-file", "", "Path to the client server TLS cert file.")
	fs.StringVar(&cfg.Ec.ClientTLSInfo.KeyFile, "key-file", "", "Path to the client server TLS key file.")
	fs.BoolVar(&cfg.Ec.ClientTLSInfo.ClientCertAuth, "client-cert-auth", false, "Enable client cert authentication.")
	fs.StringVar(&cfg.Ec.ClientTLSInfo.CRLFile, "client-crl-file", "", "Path to the client certificate revocation list file.")
	fs.StringVar(&cfg.Ec.ClientTLSInfo.TrustedCAFile, "trusted-ca-file", "", "Path to the client server TLS trusted CA cert file.")
	fs.BoolVar(&cfg.Ec.ClientAutoTLS, "auto-tls", false, "Client TLS using generated certificates")
	fs.StringVar(&cfg.Ec.PeerTLSInfo.CAFile, "peer-ca-file", "", "DEPRECATED: Path to the peer server TLS CA file.")
	fs.StringVar(&cfg.Ec.PeerTLSInfo.CertFile, "peer-cert-file", "", "Path to the peer server TLS cert file.")
	fs.StringVar(&cfg.Ec.PeerTLSInfo.KeyFile, "peer-key-file", "", "Path to the peer server TLS key file.")
	fs.BoolVar(&cfg.Ec.PeerTLSInfo.ClientCertAuth, "peer-client-cert-auth", false, "Enable peer client cert authentication.")
	fs.StringVar(&cfg.Ec.PeerTLSInfo.TrustedCAFile, "peer-trusted-ca-file", "", "Path to the peer server TLS trusted CA file.")
	fs.BoolVar(&cfg.Ec.PeerAutoTLS, "peer-auto-tls", false, "Peer TLS using generated certificates")

}

func (cfg *EtcdServerCreateConfig) ValidateFlags(cmd *cobra.Command, args []string) error {
	ensureFlags := []string{}
	//ensureFlags := []string{"config"}
	flags.EnsureRequiredFlags(cmd, ensureFlags...)

	if len(args) == 0 {
		errors.New("missing member name")
	}
	if len(args) > 1 {
		errors.New("multiple member name provided")
	}
	cfg.Name = args[0]
	cfg.Ec.Name = cfg.Name
	/*if len(cfg.PeerUrls) > 0 {
		u, err := types.NewURLs(cfg.PeerUrls)
		if err != nil {
			return fmt.Errorf("unexpected error setting up listen-metrics-urls: %v", err)
		}
		cfg.Ec.LPUrls = []url.URL(u)
	}
	if len(cfg.ClientUrls) > 0 {
		u, err := types.NewURLs(cfg.ClientUrls)
		if err != nil {
			return fmt.Errorf("unexpected error setting up listen-metrics-urls: %v", err)
		}
		cfg.Ec.LCUrls = []url.URL(u)
	}
	if len(cfg.AdvertisePeerUrls) > 0 {
		u, err := types.NewURLs(cfg.AdvertisePeerUrls)
		if err != nil {
			return fmt.Errorf("unexpected error setting up listen-metrics-urls: %v", err)
		}
		cfg.Ec.APUrls = []url.URL(u)
	}
	if len(cfg.AdvertiseClientUrls) > 0 {
		u, err := types.NewURLs(cfg.AdvertiseClientUrls)
		if err != nil {
			return fmt.Errorf("unexpected error setting up listen-metrics-urls: %v", err)
		}
		cfg.Ec.ACUrls = []url.URL(u)
	}
	*/
	if cfg.Ec.Dir == "" {
		cfg.Ec.Dir = "/tmp/" + cfg.Name
	}

	return nil
}

func (cfg *EtcdServerCreateConfig) EtcdServerConfig() etcd.ServerConfig {
	//fmt.Println(cfg.Ec.LCUrls[0].Hostname())
	return etcd.ServerConfig{
		DataDir:        cfg.Ec.Dir,
		DataQuota:      cfg.Ec.QuotaBackendBytes,
		PublicAddress:  cfg.NodeAddress,
		PrivateAddress: cfg.NodeAddress,
		ClientSC: etcd.SecurityConfig{
			CAFile:        cfg.Ec.ClientTLSInfo.CAFile,
			CertFile:      cfg.Ec.ClientTLSInfo.CertFile,
			KeyFile:       cfg.Ec.ClientTLSInfo.KeyFile,
			CertAuth:      cfg.Ec.ClientTLSInfo.ClientCertAuth,
			TrustedCAFile: cfg.Ec.ClientTLSInfo.TrustedCAFile,
			AutoTLS:       cfg.Ec.ClientAutoTLS,
		},
		PeerSC: etcd.SecurityConfig{
			CAFile:        cfg.Ec.PeerTLSInfo.CAFile,
			CertFile:      cfg.Ec.PeerTLSInfo.CertFile,
			KeyFile:       cfg.Ec.PeerTLSInfo.KeyFile,
			CertAuth:      cfg.Ec.PeerTLSInfo.ClientCertAuth,
			TrustedCAFile: cfg.Ec.PeerTLSInfo.TrustedCAFile,
			AutoTLS:       cfg.Ec.PeerAutoTLS,
		},
		UnhealthyMemberTTL: 2 * time.Minute, //cfg.UnhealthyMemberTTL,
		//SnapshotProvider:   snapshotProvider,
		//SnapshotInterval:   cfg.Snapshot.Interval,
		//SnapshotTTL:        cfg.Snapshot.TTL,

	}
}
