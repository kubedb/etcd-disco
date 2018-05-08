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

package etcdmain

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/etcd/embed"
	"github.com/ghodss/yaml"
)

func TestConfigParsingMemberFlags(t *testing.T) {
	args := []string{
		"-data-dir=testdir",
		"-name=testname",
		"-max-wals=10",
		"-max-snapshots=10",
		"-snapshot-count=10",
		"-listen-peer-urls=http://localhost:8000,https://localhost:8001",
		"-listen-client-urls=http://localhost:7000,https://localhost:7001",
		// it should be set if -listen-client-urls is set
		"-advertise-client-urls=http://localhost:7000,https://localhost:7001",
	}

	cfg := newConfig()
	err := cfg.parse(args)
	if err != nil {
		t.Fatal(err)
	}

	validateMemberFlags(t, cfg)
}

func TestConfigFileMemberFields(t *testing.T) {
	yc := struct {
		Dir           string `json:"data-dir"`
		MaxSnapFiles  uint   `json:"max-snapshots"`
		MaxWalFiles   uint   `json:"max-wals"`
		Name          string `json:"name"`
		SnapCount     uint64 `json:"snapshot-count"`
		LPUrls        string `json:"listen-peer-urls"`
		LCUrls        string `json:"listen-client-urls"`
		AcurlsCfgFile string `json:"advertise-client-urls"`
	}{
		"testdir",
		10,
		10,
		"testname",
		10,
		"http://localhost:8000,https://localhost:8001",
		"http://localhost:7000,https://localhost:7001",
		"http://localhost:7000,https://localhost:7001",
	}

	b, err := yaml.Marshal(&yc)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile := mustCreateCfgFile(t, b)
	defer os.Remove(tmpfile.Name())

	args := []string{fmt.Sprintf("--config-file=%s", tmpfile.Name())}

	cfg := newConfig()
	if err = cfg.parse(args); err != nil {
		t.Fatal(err)
	}

	validateMemberFlags(t, cfg)
}

func TestConfigParsingClusteringFlags(t *testing.T) {
	args := []string{
		"-initial-cluster=0=http://localhost:8000",
		"-initial-cluster-state=existing",
		"-initial-cluster-token=etcdtest",
		"-initial-advertise-peer-urls=http://localhost:8000,https://localhost:8001",
		"-advertise-client-urls=http://localhost:7000,https://localhost:7001",
		"-discovery-fallback=exit",
	}

	cfg := newConfig()
	if err := cfg.parse(args); err != nil {
		t.Fatal(err)
	}

	validateClusteringFlags(t, cfg)
}

func TestConfigFileClusteringFields(t *testing.T) {
	yc := struct {
		InitialCluster      string `json:"initial-cluster"`
		ClusterState        string `json:"initial-cluster-state"`
		InitialClusterToken string `json:"initial-cluster-token"`
		Apurls              string `json:"initial-advertise-peer-urls"`
		Acurls              string `json:"advertise-client-urls"`
		Fallback            string `json:"discovery-fallback"`
	}{
		"0=http://localhost:8000",
		"existing",
		"etcdtest",
		"http://localhost:8000,https://localhost:8001",
		"http://localhost:7000,https://localhost:7001",
		"exit",
	}

	b, err := yaml.Marshal(&yc)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile := mustCreateCfgFile(t, b)
	defer os.Remove(tmpfile.Name())

	args := []string{fmt.Sprintf("--config-file=%s", tmpfile.Name())}
	cfg := newConfig()
	err = cfg.parse(args)
	if err != nil {
		t.Fatal(err)
	}

	validateClusteringFlags(t, cfg)
}

func TestConfigFileClusteringFlags(t *testing.T) {
	tests := []struct {
		Name           string `json:"name"`
		InitialCluster string `json:"initial-cluster"`
		DNSCluster     string `json:"discovery-srv"`
		Durl           string `json:"discovery"`
	}{
		// Use default name and generate a default initial-cluster
		{},
		{
			Name: "non-default",
		},
		{
			InitialCluster: "0=localhost:8000",
		},
		{
			Name:           "non-default",
			InitialCluster: "0=localhost:8000",
		},
		{
			DNSCluster: "example.com",
		},
		{
			Name:       "non-default",
			DNSCluster: "example.com",
		},
		{
			Durl: "http://example.com/abc",
		},
		{
			Name: "non-default",
			Durl: "http://example.com/abc",
		},
	}

	for i, tt := range tests {
		b, err := yaml.Marshal(&tt)
		if err != nil {
			t.Fatal(err)
		}

		tmpfile := mustCreateCfgFile(t, b)
		defer os.Remove(tmpfile.Name())

		args := []string{fmt.Sprintf("--config-file=%s", tmpfile.Name())}

		cfg := newConfig()
		if err := cfg.parse(args); err != nil {
			t.Errorf("%d: err = %v", i, err)
		}
	}
}

func TestConfigParsingOtherFlags(t *testing.T) {
	args := []string{"-proxy=readonly"}

	cfg := newConfig()
	err := cfg.parse(args)
	if err != nil {
		t.Fatal(err)
	}

	validateOtherFlags(t, cfg)
}

func TestConfigFileOtherFields(t *testing.T) {
	yc := struct {
		ProxyCfgFile string `json:"proxy"`
	}{
		"readonly",
	}

	b, err := yaml.Marshal(&yc)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile := mustCreateCfgFile(t, b)
	defer os.Remove(tmpfile.Name())

	args := []string{fmt.Sprintf("--config-file=%s", tmpfile.Name())}

	cfg := newConfig()
	err = cfg.parse(args)
	if err != nil {
		t.Fatal(err)
	}

	validateOtherFlags(t, cfg)
}

func TestConfigParsingConflictClusteringFlags(t *testing.T) {
	conflictArgs := [][]string{
		{
			"-initial-cluster=0=localhost:8000",
			"-discovery=http://example.com/abc",
		},
		{
			"-discovery-srv=example.com",
			"-discovery=http://example.com/abc",
		},
		{
			"-initial-cluster=0=localhost:8000",
			"-discovery-srv=example.com",
		},
		{
			"-initial-cluster=0=localhost:8000",
			"-discovery=http://example.com/abc",
			"-discovery-srv=example.com",
		},
	}

	for i, tt := range conflictArgs {
		cfg := newConfig()
		if err := cfg.parse(tt); err != embed.ErrConflictBootstrapFlags {
			t.Errorf("%d: err = %v, want %v", i, err, embed.ErrConflictBootstrapFlags)
		}
	}
}

func TestConfigFileConflictClusteringFlags(t *testing.T) {
	tests := []struct {
		InitialCluster string `json:"initial-cluster"`
		DNSCluster     string `json:"discovery-srv"`
		Durl           string `json:"discovery"`
	}{
		{
			InitialCluster: "0=localhost:8000",
			Durl:           "http://example.com/abc",
		},
		{
			DNSCluster: "example.com",
			Durl:       "http://example.com/abc",
		},
		{
			InitialCluster: "0=localhost:8000",
			DNSCluster:     "example.com",
		},
		{
			InitialCluster: "0=localhost:8000",
			Durl:           "http://example.com/abc",
			DNSCluster:     "example.com",
		},
	}

	for i, tt := range tests {
		b, err := yaml.Marshal(&tt)
		if err != nil {
			t.Fatal(err)
		}

		tmpfile := mustCreateCfgFile(t, b)
		defer os.Remove(tmpfile.Name())

		args := []string{fmt.Sprintf("--config-file=%s", tmpfile.Name())}

		cfg := newConfig()
		if err := cfg.parse(args); err != embed.ErrConflictBootstrapFlags {
			t.Errorf("%d: err = %v, want %v", i, err, embed.ErrConflictBootstrapFlags)
		}
	}
}

func TestConfigParsingMissedAdvertiseClientURLsFlag(t *testing.T) {
	tests := []struct {
		args []string
		werr error
	}{
		{
			[]string{
				"-initial-cluster=infra1=http://127.0.0.1:2380",
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			embed.ErrUnsetAdvertiseClientURLsFlag,
		},
		{
			[]string{
				"-discovery-srv=example.com",
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			embed.ErrUnsetAdvertiseClientURLsFlag,
		},
		{
			[]string{
				"-discovery=http://example.com/abc",
				"-discovery-fallback=exit",
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			embed.ErrUnsetAdvertiseClientURLsFlag,
		},
		{
			[]string{
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			embed.ErrUnsetAdvertiseClientURLsFlag,
		},
		{
			[]string{
				"-discovery=http://example.com/abc",
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			nil,
		},
		{
			[]string{
				"-proxy=on",
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			nil,
		},
		{
			[]string{
				"-proxy=readonly",
				"-listen-client-urls=http://127.0.0.1:2379",
			},
			nil,
		},
	}

	for i, tt := range tests {
		cfg := newConfig()
		if err := cfg.parse(tt.args); err != tt.werr {
			t.Errorf("%d: err = %v, want %v", i, err, tt.werr)
		}
	}
}

func TestConfigIsNewCluster(t *testing.T) {
	tests := []struct {
		state  string
		wIsNew bool
	}{
		{embed.ClusterStateFlagExisting, false},
		{embed.ClusterStateFlagNew, true},
	}
	for i, tt := range tests {
		cfg := newConfig()
		args := []string{"--initial-cluster-state", tests[i].state}
		if err := cfg.parse(args); err != nil {
			t.Fatalf("#%d: unexpected clusterState.Set error: %v", i, err)
		}
		if g := cfg.Ec.IsNewCluster(); g != tt.wIsNew {
			t.Errorf("#%d: isNewCluster = %v, want %v", i, g, tt.wIsNew)
		}
	}
}

func TestConfigIsProxy(t *testing.T) {
	tests := []struct {
		proxy    string
		wIsProxy bool
	}{
		{proxyFlagOff, false},
		{proxyFlagReadonly, true},
		{proxyFlagOn, true},
	}
	for i, tt := range tests {
		cfg := newConfig()
		if err := cfg.Cf.proxy.Set(tt.proxy); err != nil {
			t.Fatalf("#%d: unexpected proxy.Set error: %v", i, err)
		}
		if g := cfg.isProxy(); g != tt.wIsProxy {
			t.Errorf("#%d: isProxy = %v, want %v", i, g, tt.wIsProxy)
		}
	}
}

func TestConfigIsReadonlyProxy(t *testing.T) {
	tests := []struct {
		proxy       string
		wIsReadonly bool
	}{
		{proxyFlagOff, false},
		{proxyFlagReadonly, true},
		{proxyFlagOn, false},
	}
	for i, tt := range tests {
		cfg := newConfig()
		if err := cfg.Cf.proxy.Set(tt.proxy); err != nil {
			t.Fatalf("#%d: unexpected proxy.Set error: %v", i, err)
		}
		if g := cfg.isReadonlyProxy(); g != tt.wIsReadonly {
			t.Errorf("#%d: isReadonlyProxy = %v, want %v", i, g, tt.wIsReadonly)
		}
	}
}

func TestConfigShouldFallbackToProxy(t *testing.T) {
	tests := []struct {
		fallback  string
		wFallback bool
	}{
		{fallbackFlagProxy, true},
		{fallbackFlagExit, false},
	}
	for i, tt := range tests {
		cfg := newConfig()
		if err := cfg.Cf.fallback.Set(tt.fallback); err != nil {
			t.Fatalf("#%d: unexpected fallback.Set error: %v", i, err)
		}
		if g := cfg.shouldFallbackToProxy(); g != tt.wFallback {
			t.Errorf("#%d: shouldFallbackToProxy = %v, want %v", i, g, tt.wFallback)
		}
	}
}

func TestConfigFileElectionTimeout(t *testing.T) {
	tests := []struct {
		TickMs     uint `json:"heartbeat-interval"`
		ElectionMs uint `json:"election-timeout"`
		errStr     string
	}{
		{
			ElectionMs: 1000,
			TickMs:     800,
			errStr:     "should be at least as 5 times as",
		},
		{
			ElectionMs: 60000,
			TickMs:     10000,
			errStr:     "is too long, and should be set less than",
		},
		{
			ElectionMs: 100,
			TickMs:     0,
			errStr:     "--heartbeat-interval must be >0 (set to 0ms)",
		},
		{
			ElectionMs: 0,
			TickMs:     100,
			errStr:     "--election-timeout must be >0 (set to 0ms)",
		},
	}

	for i, tt := range tests {
		b, err := yaml.Marshal(&tt)
		if err != nil {
			t.Fatal(err)
		}

		tmpfile := mustCreateCfgFile(t, b)
		defer os.Remove(tmpfile.Name())

		args := []string{fmt.Sprintf("--config-file=%s", tmpfile.Name())}

		cfg := newConfig()
		if err := cfg.parse(args); err == nil || !strings.Contains(err.Error(), tt.errStr) {
			t.Errorf("%d: Wrong err = %v", i, err)
		}
	}
}

func mustCreateCfgFile(t *testing.T, b []byte) *os.File {
	tmpfile, err := ioutil.TempFile("", "servercfg")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmpfile.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return tmpfile
}

func validateMemberFlags(t *testing.T, cfg *Config) {
	wcfg := &embed.Config{
		Dir:          "testdir",
		LPUrls:       []url.URL{{Scheme: "http", Host: "localhost:8000"}, {Scheme: "https", Host: "localhost:8001"}},
		LCUrls:       []url.URL{{Scheme: "http", Host: "localhost:7000"}, {Scheme: "https", Host: "localhost:7001"}},
		MaxSnapFiles: 10,
		MaxWalFiles:  10,
		Name:         "testname",
		SnapCount:    10,
	}

	if cfg.Ec.Dir != wcfg.Dir {
		t.Errorf("dir = %v, want %v", cfg.Ec.Dir, wcfg.Dir)
	}
	if cfg.Ec.MaxSnapFiles != wcfg.MaxSnapFiles {
		t.Errorf("maxsnap = %v, want %v", cfg.Ec.MaxSnapFiles, wcfg.MaxSnapFiles)
	}
	if cfg.Ec.MaxWalFiles != wcfg.MaxWalFiles {
		t.Errorf("maxwal = %v, want %v", cfg.Ec.MaxWalFiles, wcfg.MaxWalFiles)
	}
	if cfg.Ec.Name != wcfg.Name {
		t.Errorf("name = %v, want %v", cfg.Ec.Name, wcfg.Name)
	}
	if cfg.Ec.SnapCount != wcfg.SnapCount {
		t.Errorf("snapcount = %v, want %v", cfg.Ec.SnapCount, wcfg.SnapCount)
	}
	if !reflect.DeepEqual(cfg.Ec.LPUrls, wcfg.LPUrls) {
		t.Errorf("listen-peer-urls = %v, want %v", cfg.Ec.LPUrls, wcfg.LPUrls)
	}
	if !reflect.DeepEqual(cfg.Ec.LCUrls, wcfg.LCUrls) {
		t.Errorf("listen-client-urls = %v, want %v", cfg.Ec.LCUrls, wcfg.LCUrls)
	}
}

func validateClusteringFlags(t *testing.T, cfg *Config) {
	wcfg := newConfig()
	wcfg.Ec.APUrls = []url.URL{{Scheme: "http", Host: "localhost:8000"}, {Scheme: "https", Host: "localhost:8001"}}
	wcfg.Ec.ACUrls = []url.URL{{Scheme: "http", Host: "localhost:7000"}, {Scheme: "https", Host: "localhost:7001"}}
	wcfg.Ec.ClusterState = embed.ClusterStateFlagExisting
	wcfg.Cf.fallback.Set(fallbackFlagExit)
	wcfg.Ec.InitialCluster = "0=http://localhost:8000"
	wcfg.Ec.InitialClusterToken = "etcdtest"

	if cfg.Ec.ClusterState != wcfg.Ec.ClusterState {
		t.Errorf("clusterState = %v, want %v", cfg.Ec.ClusterState, wcfg.Ec.ClusterState)
	}
	if cfg.Cf.fallback.String() != wcfg.Cf.fallback.String() {
		t.Errorf("fallback = %v, want %v", cfg.Cf.fallback, wcfg.Cf.fallback)
	}
	if cfg.Ec.InitialCluster != wcfg.Ec.InitialCluster {
		t.Errorf("initialCluster = %v, want %v", cfg.Ec.InitialCluster, wcfg.Ec.InitialCluster)
	}
	if cfg.Ec.InitialClusterToken != wcfg.Ec.InitialClusterToken {
		t.Errorf("initialClusterToken = %v, want %v", cfg.Ec.InitialClusterToken, wcfg.Ec.InitialClusterToken)
	}
	if !reflect.DeepEqual(cfg.Ec.APUrls, wcfg.Ec.APUrls) {
		t.Errorf("initial-advertise-peer-urls = %v, want %v", cfg.Ec.LPUrls, wcfg.Ec.LPUrls)
	}
	if !reflect.DeepEqual(cfg.Ec.ACUrls, wcfg.Ec.ACUrls) {
		t.Errorf("advertise-client-urls = %v, want %v", cfg.Ec.LCUrls, wcfg.Ec.LCUrls)
	}
}

func validateOtherFlags(t *testing.T, cfg *Config) {
	wcfg := newConfig()
	wcfg.Cf.proxy.Set(proxyFlagReadonly)
	if cfg.Cf.proxy.String() != wcfg.Cf.proxy.String() {
		t.Errorf("proxy = %v, want %v", cfg.Cf.proxy, wcfg.Cf.proxy)
	}
}
