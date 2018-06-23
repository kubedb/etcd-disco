package pkg

import (
	"github.com/appscode/etcd-disco/pkg/etcd"
	"github.com/appscode/etcd-disco/pkg/etcdmain"
)

func EtcdServerConfig(cfg *etcdmain.Config) etcd.ServerConfig {
	return etcd.ServerConfig{
		DataDir:        cfg.Ec.Dir,
		DataQuota:      cfg.Ec.QuotaBackendBytes,
		PublicAddress:  cfg.Ec.LCUrls[0].Host,
		PrivateAddress: cfg.Ec.LCUrls[0].Host,
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
		//UnhealthyMemberTTL: cfg.UnhealthyMemberTTL,
		//SnapshotProvider:   snapshotProvider,
		//SnapshotInterval:   cfg.Snapshot.Interval,
		//SnapshotTTL:        cfg.Snapshot.TTL,

	}
}
