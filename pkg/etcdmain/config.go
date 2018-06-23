// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Every change should be reflected on help.go as well.

package etcdmain

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/pkg/flags"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/version"
	"github.com/ghodss/yaml"
)

var (
	proxyFlagOff      = "off"
	proxyFlagReadonly = "readonly"
	proxyFlagOn       = "on"

	fallbackFlagExit  = "exit"
	fallbackFlagProxy = "proxy"

	ignored = []string{
		"cluster-active-size",
		"cluster-remove-delay",
		"cluster-sync-interval",
		"config",
		"force",
		"max-result-buffer",
		"max-retry-attempts",
		"peer-heartbeat-interval",
		"peer-election-timeout",
		"retry-interval",
		"snapshot",
		"v",
		"vv",
		// for coverage testing
		"test.coverprofile",
		"test.outputdir",
	}
)

type configProxy struct {
	ProxyFailureWaitMs     uint `json:"proxy-failure-wait"`
	ProxyRefreshIntervalMs uint `json:"proxy-refresh-interval"`
	ProxyDialTimeoutMs     uint `json:"proxy-dial-timeout"`
	ProxyWriteTimeoutMs    uint `json:"proxy-write-timeout"`
	ProxyReadTimeoutMs     uint `json:"proxy-read-timeout"`
	Fallback               string
	Proxy                  string
	ProxyJSON              string `json:"proxy"`
	FallbackJSON           string `json:"discovery-fallback"`
}

// Config holds the Config for a command line invocation of etcd
type Config struct {
	Ec           embed.Config
	Cp           configProxy
	Cf           configFlags
	ConfigFile   string
	PrintVersion bool
	Ignored      []string
}

// configFlags has the set of flags used for command line parsing a Config
type configFlags struct {
	FlagSet      *flag.FlagSet
	clusterState *flags.StringsFlag
	fallback     *flags.StringsFlag
	proxy        *flags.StringsFlag
}

func NewConfig() *Config {
	cfg := &Config{
		Ec: *embed.NewConfig(),
		Cp: configProxy{
			Proxy:                  proxyFlagOff,
			ProxyFailureWaitMs:     5000,
			ProxyRefreshIntervalMs: 30000,
			ProxyDialTimeoutMs:     1000,
			ProxyWriteTimeoutMs:    5000,
		},
		Ignored: ignored,
	}
	cfg.Cf = configFlags{
		FlagSet: flag.NewFlagSet("etcd", flag.ContinueOnError),
		clusterState: flags.NewStringsFlag(
			embed.ClusterStateFlagNew,
			embed.ClusterStateFlagExisting,
		),
		fallback: flags.NewStringsFlag(
			fallbackFlagProxy,
			fallbackFlagExit,
		),
		proxy: flags.NewStringsFlag(
			proxyFlagOff,
			proxyFlagReadonly,
			proxyFlagOn,
		),
	}

	fs := cfg.Cf.FlagSet
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, usageline)
	}

	fs.StringVar(&cfg.ConfigFile, "config-file", "", "Path to the server configuration file")

	// member
	fs.Var(cfg.Ec.CorsInfo, "cors", "Comma-separated white list of origins for CORS (cross-origin resource sharing).")
	fs.StringVar(&cfg.Ec.Dir, "data-dir", cfg.Ec.Dir, "Path to the data directory.")
	fs.StringVar(&cfg.Ec.WalDir, "wal-dir", cfg.Ec.WalDir, "Path to the dedicated wal directory.")
	fs.Var(flags.NewURLsValue(embed.DefaultListenPeerURLs), "listen-peer-urls", "List of URLs to listen on for peer traffic.")
	fs.Var(flags.NewURLsValue(embed.DefaultListenClientURLs), "listen-client-urls", "List of URLs to listen on for client traffic.")
	fs.StringVar(&cfg.Ec.ListenMetricsUrlsJSON, "listen-metrics-urls", "", "List of URLs to listen on for metrics.")
	fs.UintVar(&cfg.Ec.MaxSnapFiles, "max-snapshots", cfg.Ec.MaxSnapFiles, "Maximum number of snapshot files to retain (0 is unlimited).")
	fs.UintVar(&cfg.Ec.MaxWalFiles, "max-wals", cfg.Ec.MaxWalFiles, "Maximum number of wal files to retain (0 is unlimited).")
	fs.StringVar(&cfg.Ec.Name, "name", cfg.Ec.Name, "Human-readable name for this member.")
	fs.Uint64Var(&cfg.Ec.SnapCount, "snapshot-count", cfg.Ec.SnapCount, "Number of committed transactions to trigger a snapshot to disk.")
	fs.UintVar(&cfg.Ec.TickMs, "heartbeat-interval", cfg.Ec.TickMs, "Time (in milliseconds) of a heartbeat interval.")
	fs.UintVar(&cfg.Ec.ElectionMs, "election-timeout", cfg.Ec.ElectionMs, "Time (in milliseconds) for an election to timeout.")
	fs.Int64Var(&cfg.Ec.QuotaBackendBytes, "quota-backend-bytes", cfg.Ec.QuotaBackendBytes, "Raise alarms when backend size exceeds the given quota. 0 means use the default quota.")
	fs.UintVar(&cfg.Ec.MaxTxnOps, "max-txn-ops", cfg.Ec.MaxTxnOps, "Maximum number of operations permitted in a transaction.")
	fs.UintVar(&cfg.Ec.MaxRequestBytes, "max-request-bytes", cfg.Ec.MaxRequestBytes, "Maximum client request size in bytes the server will accept.")
	fs.DurationVar(&cfg.Ec.GRPCKeepAliveMinTime, "grpc-keepalive-min-time", cfg.Ec.GRPCKeepAliveMinTime, "Minimum interval duration that a client should wait before pinging server.")
	fs.DurationVar(&cfg.Ec.GRPCKeepAliveInterval, "grpc-keepalive-interval", cfg.Ec.GRPCKeepAliveInterval, "Frequency duration of server-to-client ping to check if a connection is alive (0 to disable).")
	fs.DurationVar(&cfg.Ec.GRPCKeepAliveTimeout, "grpc-keepalive-timeout", cfg.Ec.GRPCKeepAliveTimeout, "Additional duration of wait before closing a non-responsive connection (0 to disable).")

	// clustering
	fs.Var(flags.NewURLsValue(embed.DefaultInitialAdvertisePeerURLs), "initial-advertise-peer-urls", "List of this member's peer URLs to advertise to the rest of the cluster.")
	fs.Var(flags.NewURLsValue(embed.DefaultAdvertiseClientURLs), "advertise-client-urls", "List of this member's client URLs to advertise to the public.")
	fs.StringVar(&cfg.Ec.Durl, "discovery", cfg.Ec.Durl, "Discovery URL used to bootstrap the cluster.")
	fs.Var(cfg.Cf.fallback, "discovery-fallback", fmt.Sprintf("Valid values include %s", strings.Join(cfg.Cf.fallback.Values, ", ")))

	fs.StringVar(&cfg.Ec.Dproxy, "discovery-proxy", cfg.Ec.Dproxy, "HTTP proxy to use for traffic to discovery service.")
	fs.StringVar(&cfg.Ec.DNSCluster, "discovery-srv", cfg.Ec.DNSCluster, "DNS domain used to bootstrap initial cluster.")
	fs.StringVar(&cfg.Ec.InitialCluster, "initial-cluster", cfg.Ec.InitialCluster, "Initial cluster configuration for bootstrapping.")
	fs.StringVar(&cfg.Ec.InitialClusterToken, "initial-cluster-token", cfg.Ec.InitialClusterToken, "Initial cluster token for the etcd cluster during bootstrap.")
	fs.Var(cfg.Cf.clusterState, "initial-cluster-state", "Initial cluster state ('new' or 'existing').")

	fs.BoolVar(&cfg.Ec.StrictReconfigCheck, "strict-reconfig-check", cfg.Ec.StrictReconfigCheck, "Reject reconfiguration requests that would cause quorum loss.")
	fs.BoolVar(&cfg.Ec.EnableV2, "enable-v2", cfg.Ec.EnableV2, "Accept etcd V2 client requests.")
	fs.StringVar(&cfg.Ec.ExperimentalEnableV2V3, "experimental-enable-v2v3", cfg.Ec.ExperimentalEnableV2V3, "v3 prefix for serving emulated v2 state.")

	// proxy
	fs.Var(cfg.Cf.proxy, "proxy", fmt.Sprintf("Valid values include %s", strings.Join(cfg.Cf.proxy.Values, ", ")))

	fs.UintVar(&cfg.Cp.ProxyFailureWaitMs, "proxy-failure-wait", cfg.Cp.ProxyFailureWaitMs, "Time (in milliseconds) an endpoint will be held in a failed state.")
	fs.UintVar(&cfg.Cp.ProxyRefreshIntervalMs, "proxy-refresh-interval", cfg.Cp.ProxyRefreshIntervalMs, "Time (in milliseconds) of the endpoints refresh interval.")
	fs.UintVar(&cfg.Cp.ProxyDialTimeoutMs, "proxy-dial-timeout", cfg.Cp.ProxyDialTimeoutMs, "Time (in milliseconds) for a dial to timeout.")
	fs.UintVar(&cfg.Cp.ProxyWriteTimeoutMs, "proxy-write-timeout", cfg.Cp.ProxyWriteTimeoutMs, "Time (in milliseconds) for a write to timeout.")
	fs.UintVar(&cfg.Cp.ProxyReadTimeoutMs, "proxy-read-timeout", cfg.Cp.ProxyReadTimeoutMs, "Time (in milliseconds) for a read to timeout.")

	// security
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
	fs.StringVar(&cfg.Ec.PeerTLSInfo.CRLFile, "peer-crl-file", "", "Path to the peer certificate revocation list file.")
	fs.StringVar(&cfg.Ec.PeerTLSInfo.AllowedCN, "peer-cert-allowed-cn", "", "Allowed CN for inter peer authentication.")

	// logging
	fs.BoolVar(&cfg.Ec.Debug, "debug", false, "Enable debug-level logging for etcd.")
	fs.StringVar(&cfg.Ec.LogPkgLevels, "log-package-levels", "", "Specify a particular log level for each etcd package (eg: 'etcdmain=CRITICAL,etcdserver=DEBUG').")
	fs.StringVar(&cfg.Ec.LogOutput, "log-output", embed.DefaultLogOutput, "Specify 'stdout' or 'stderr' to skip journald logging even when running under systemd.")

	// unsafe
	fs.BoolVar(&cfg.Ec.ForceNewCluster, "force-new-cluster", false, "Force to create a new one member cluster.")

	// version
	fs.BoolVar(&cfg.PrintVersion, "version", false, "Print the version and exit.")

	fs.StringVar(&cfg.Ec.AutoCompactionRetention, "auto-compaction-retention", "0", "Auto compaction retention for mvcc key value store. 0 means disable auto compaction.")
	fs.StringVar(&cfg.Ec.AutoCompactionMode, "auto-compaction-mode", "periodic", "interpret 'auto-compaction-retention' one of: periodic|revision. 'periodic' for duration based retention, defaulting to hours if no time unit is provided (e.g. '5m'). 'revision' for revision number based retention.")

	// pprof profiler via HTTP
	fs.BoolVar(&cfg.Ec.EnablePprof, "enable-pprof", false, "Enable runtime profiling data via HTTP server. Address is at client URL + \"/debug/pprof/\"")

	// additional metrics
	fs.StringVar(&cfg.Ec.Metrics, "metrics", cfg.Ec.Metrics, "Set level of detail for exported metrics, specify 'extensive' to include histogram metrics")

	// auth
	fs.StringVar(&cfg.Ec.AuthToken, "auth-token", cfg.Ec.AuthToken, "Specify auth token specific options.")

	// experimental
	fs.BoolVar(&cfg.Ec.ExperimentalInitialCorruptCheck, "experimental-initial-corrupt-check", cfg.Ec.ExperimentalInitialCorruptCheck, "Enable to check data corruption before serving any client/peer traffic.")
	fs.DurationVar(&cfg.Ec.ExperimentalCorruptCheckTime, "experimental-corrupt-check-time", cfg.Ec.ExperimentalCorruptCheckTime, "Duration of time between cluster corruption check passes.")

	// Ignored
	for _, f := range cfg.Ignored {
		fs.Var(&flags.IgnoredFlag{Name: f}, f, "")
	}
	return cfg
}

func (cfg *Config) Parse(arguments []string) error {
	perr := cfg.Cf.FlagSet.Parse(arguments)
	switch perr {
	case nil:
	case flag.ErrHelp:
		fmt.Println(flagsline)
		os.Exit(0)
	default:
		os.Exit(2)
	}
	if len(cfg.Cf.FlagSet.Args()) != 0 {
		return fmt.Errorf("'%s' is not a valid flag", cfg.Cf.FlagSet.Arg(0))
	}

	if cfg.PrintVersion {
		fmt.Printf("etcd Version: %s\n", version.Version)
		fmt.Printf("Git SHA: %s\n", version.GitSHA)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	var err error
	if cfg.ConfigFile != "" {
		plog.Infof("Loading server configuration from %q", cfg.ConfigFile)
		err = cfg.configFromFile(cfg.ConfigFile)
	} else {
		err = cfg.ConfigFromCmdLine()
	}
	return err
}

func (cfg *Config) ConfigFromCmdLine() error {
	err := flags.SetFlagsFromEnv("ETCD", cfg.Cf.FlagSet)
	if err != nil {
		plog.Fatalf("%v", err)
	}

	cfg.Ec.LPUrls = flags.URLsFromFlag(cfg.Cf.FlagSet, "listen-peer-urls")
	cfg.Ec.APUrls = flags.URLsFromFlag(cfg.Cf.FlagSet, "initial-advertise-peer-urls")
	cfg.Ec.LCUrls = flags.URLsFromFlag(cfg.Cf.FlagSet, "listen-client-urls")
	cfg.Ec.ACUrls = flags.URLsFromFlag(cfg.Cf.FlagSet, "advertise-client-urls")

	if len(cfg.Ec.ListenMetricsUrlsJSON) > 0 {
		u, err := types.NewURLs(strings.Split(cfg.Ec.ListenMetricsUrlsJSON, ","))
		if err != nil {
			plog.Fatalf("unexpected error setting up listen-metrics-urls: %v", err)
		}
		cfg.Ec.ListenMetricsUrls = []url.URL(u)
	}

	cfg.Ec.ClusterState = cfg.Cf.clusterState.String()
	cfg.Cp.Fallback = cfg.Cf.fallback.String()
	cfg.Cp.Proxy = cfg.Cf.proxy.String()

	// disable default advertise-client-urls if lcurls is set
	missingAC := flags.IsSet(cfg.Cf.FlagSet, "listen-client-urls") && !flags.IsSet(cfg.Cf.FlagSet, "advertise-client-urls")
	if !cfg.mayBeProxy() && missingAC {
		cfg.Ec.ACUrls = nil
	}

	// disable default initial-cluster if discovery is set
	if (cfg.Ec.Durl != "" || cfg.Ec.DNSCluster != "") && !flags.IsSet(cfg.Cf.FlagSet, "initial-cluster") {
		cfg.Ec.InitialCluster = ""
	}

	return cfg.validate()
}

func (cfg *Config) configFromFile(path string) error {
	eCfg, err := embed.ConfigFromFile(path)
	if err != nil {
		return err
	}
	cfg.Ec = *eCfg

	// load extra Config information
	b, rerr := ioutil.ReadFile(path)
	if rerr != nil {
		return rerr
	}
	if yerr := yaml.Unmarshal(b, &cfg.Cp); yerr != nil {
		return yerr
	}
	if cfg.Cp.FallbackJSON != "" {
		if err := cfg.Cf.fallback.Set(cfg.Cp.FallbackJSON); err != nil {
			plog.Panicf("unexpected error setting up discovery-fallback flag: %v", err)
		}
		cfg.Cp.Fallback = cfg.Cf.fallback.String()
	}
	if cfg.Cp.ProxyJSON != "" {
		if err := cfg.Cf.proxy.Set(cfg.Cp.ProxyJSON); err != nil {
			plog.Panicf("unexpected error setting up proxyFlag: %v", err)
		}
		cfg.Cp.Proxy = cfg.Cf.proxy.String()
	}
	return nil
}

func (cfg *Config) mayBeProxy() bool {
	mayFallbackToProxy := cfg.Ec.Durl != "" && cfg.Cp.Fallback == fallbackFlagProxy
	return cfg.Cp.Proxy != proxyFlagOff || mayFallbackToProxy
}

func (cfg *Config) validate() error {
	err := cfg.Ec.Validate()
	// TODO(yichengq): check this for joining through discovery service case
	if err == embed.ErrUnsetAdvertiseClientURLsFlag && cfg.mayBeProxy() {
		return nil
	}
	return err
}

func (cfg Config) isProxy() bool               { return cfg.Cf.proxy.String() != proxyFlagOff }
func (cfg Config) isReadonlyProxy() bool       { return cfg.Cf.proxy.String() == proxyFlagReadonly }
func (cfg Config) shouldFallbackToProxy() bool { return cfg.Cf.fallback.String() == fallbackFlagProxy }
