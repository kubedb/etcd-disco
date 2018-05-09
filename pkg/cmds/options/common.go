package options

import (
	"time"

	"github.com/appscode/go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	//	"github.com/pkg/errors"
	"github.com/etcd-manager/lector/pkg/etcd"
)

type EtcdClusterConfig struct {
	*etcd.ServerConfig
}

func NewEtcdClusterConfig() *EtcdClusterConfig {
	return &EtcdClusterConfig{
		&etcd.ServerConfig{
			CheckInterval:        15 * time.Second,
			UnhealthyMemberTTL:   30 * time.Second,
			AutoDisasterRecovery: true,

			DiscoveryFile:    "",
			SnapshotDir:      "",
			SnapshotInterval: 2 * time.Minute,
		},
	}
}

func (cfg *EtcdClusterConfig) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&cfg.CheckInterval, "check-interval", cfg.CheckInterval, "The interval between each cluster verification by the operator.")
	fs.DurationVar(&cfg.UnhealthyMemberTTL, "unhealthy-member-ttl", cfg.UnhealthyMemberTTL, "The time after which, an unhealthy member will be removed from the cluster.")
	fs.BoolVar(&cfg.AutoDisasterRecovery, "auto-disaster-recovery", cfg.AutoDisasterRecovery, "Defines whether the operator will attempt to seed a new cluster from a snapshot after the managed cluster has lost quorum")

	fs.StringVar(&cfg.DiscoveryFile, "discovery-file", cfg.DiscoveryFile, "Disovery file location")
	fs.StringVar(&cfg.SnapshotDir, "snapshot-dir", cfg.SnapshotDir, "Snapshot directory location")
}

func (cfg *EtcdClusterConfig) ValidateFlags(cmd *cobra.Command, args []string) error {
	ensureFlags := []string{"name"}
	flags.EnsureRequiredFlags(cmd, ensureFlags...)
	name := cmd.Flag("name").Value.String()
	if cfg.DiscoveryFile == "" {
		cfg.DiscoveryFile = "/tmp/etcd/discovery/" + name
	}
	if cfg.SnapshotDir == "" {
		cfg.SnapshotDir = "/tmp/etcd/snapshot/" + name
	}
	return nil

}
