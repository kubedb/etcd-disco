package config

import (
	"github.com/etcd-manager/lector/pkg/etcd"
	//"github.com/quentin-m/etcd-cloud-operator/pkg/providers/snapshot"
	"os"
	"io/ioutil"
	"github.com/ghodss/yaml"
)

type Config struct {
	Etcd etcd.EtcdConfiguration `yaml:"etcd"`
	//Snapshot snapshot.Config
}

func DefaultConfig() *Config {
	return &Config{
		Etcd: etcd.EtcdConfiguration{
			BackendQuota: 2 * 1024 * 1024 * 1024,
			PeerTransportSecurity: etcd.SecurityConfig{
				AutoTLS: true,
			},
		},
	}
}

func LoadConfig(configPath string) (*Config, error) {
	conf := DefaultConfig()
	if configPath == "" {
		return conf, nil
	}
	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}