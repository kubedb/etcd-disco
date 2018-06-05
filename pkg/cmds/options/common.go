package options

import (
	"time"

	"github.com/appscode/go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	//	"github.com/pkg/errors"
	"fmt"

	"github.com/appscode/go/net"
	"github.com/etcd-manager/lector/pkg/etcd"
)

const (
	ClusterTypeSeed = "seed"
	ClusterTypeJoin = "join"
)

type EtcdClusterConfig struct {
	*etcd.ServerConfig
	//ClusterType   string
	ServerAddress []string
	SelfAddrss    string
	//ClusterSize   int
}

func NewEtcdClusterConfig() *EtcdClusterConfig {
	return &EtcdClusterConfig{
		&etcd.ServerConfig{
			CheckInterval:        30 * time.Minute,
			UnhealthyMemberTTL:   2 * time.Minute,
			AutoDisasterRecovery: true,

			SnapshotInterval: 24 * time.Hour,
		},
		//ClusterTypeSeed,
		[]string{},
		"",
		//	1,
	}
}

func (cfg *EtcdClusterConfig) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&cfg.CheckInterval, "check-interval", cfg.CheckInterval, "The interval between each cluster verification by the operator.")
	fs.DurationVar(&cfg.UnhealthyMemberTTL, "unhealthy-member-ttl", cfg.UnhealthyMemberTTL, "The time after which, an unhealthy member will be removed from the cluster.")
	fs.BoolVar(&cfg.AutoDisasterRecovery, "auto-disaster-recovery", cfg.AutoDisasterRecovery, "Defines whether the operator will attempt to seed a new cluster from a snapshot after the managed cluster has lost quorum")

	//fs.StringVar(&cfg.ClusterType, "cluster-type", cfg.ClusterType, "cluster type (seed/join)")
	fs.StringArrayVar(&cfg.ServerAddress, "server-address", cfg.ServerAddress, "List of URLs to listen on for peer traffic. (required for join)")
	//fs.StringVar(&cfg.SelfAddrss, "self-address", cfg.SelfAddrss, "Own address")
	//	fs.IntVar(&cfg.ClusterSize, "cluster-size", cfg.ClusterSize, "Size of actual cluster(number of master)")
	//fs.StringVar(&cfg.DiscoveryFile, "discovery-file", cfg.DiscoveryFile, "Disovery file location")
	//fs.StringVar(&cfg.SnapshotDir, "snapshot-dir", cfg.SnapshotDir, "Snapshot directory location")
}

func (cfg *EtcdClusterConfig) ValidateFlags(cmd *cobra.Command, args []string) error {
	ensureFlags := []string{"name"}
	flags.EnsureRequiredFlags(cmd, ensureFlags...)
	ips, _, err := net.RoutableIPs()
	if err != nil {
		return fmt.Errorf("failed to detect routable ips. Reason: %v", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("no routable ips found")
	}
	fmt.Println(ips)
	cfg.SelfAddrss = ips[0]
	fmt.Println("found self address = ", cfg.SelfAddrss, "**************")
	return nil

}
