// Copyright 2017 Quentin Machu & eco authors
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

package operator

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/appscode/etcd-disco/pkg/etcd"
	"github.com/appscode/etcd-disco/pkg/etcdmain"
	"github.com/appscode/etcd-disco/pkg/providers/snapshot"
	log "github.com/sirupsen/logrus"
)

const (
	loopInterval = 15 * time.Second

	webServerPort = 2378
)

type Operator struct {
	server *etcd.Server

	// New()
	cfg Config
	//asgProvider      asg.Provider
	snapshotProvider snapshot.Provider
	initialInstances []string

	httpClient *http.Client

	shutdownChan chan os.Signal
	shutdown     bool
	ticker       *time.Ticker

	// evaluate()
	etcdHealthy bool
	etcdRunning bool

	etcdClient   *etcd.Client
	etcdSnapshot *snapshot.Metadata

	state  string
	states map[string]int

	isSeeder    bool
	clusterSize int
}

// Config is the global configuration for an instance of ECO.
type Config struct {
	UnhealthyMemberTTL time.Duration `yaml:"unhealthy-member-ttl"`

	Etcd *etcdmain.Config/*EtcdConfiguration*/ `yaml:"etcd"`
	//ASG      asg.Config             `yaml:"asg"`
	Snapshot                snapshot.Config `yaml:"snapshot"`
	InitialMembersAddresses []string        `yaml:"initial-member-addresses"`
	CurrentMemberAddress    string          `yaml:"current-member-addres"`
	//ClusterSize             int             `yaml:"custer-size"`
}

func New(cfg Config) *Operator {
	// Initialize providers.
	/*asgProvider, snapshotProvider := initProviders(cfg)
	if snapshotProvider == nil || cfg.Snapshot.Interval == 0 {
		log.Fatal("snapshots must be enabled for auto disaster recovery")
	}*/
	snapshotProvider := initSnapshotProvider(cfg.Snapshot)

	// Setup signal handler.
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM)
	var httpClient *http.Client
	if cfg.Etcd.Ec.PeerAutoTLS || !cfg.Etcd.Ec.PeerTLSInfo.Empty() {
		cert, err := tls.LoadX509KeyPair(cfg.Etcd.Ec.PeerTLSInfo.CertFile, cfg.Etcd.Ec.PeerTLSInfo.KeyFile)
		fmt.Println(err)
		caCert, err := ioutil.ReadFile(cfg.Etcd.Ec.PeerTLSInfo.TrustedCAFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      caCertPool,
			},
		}
		httpClient = &http.Client{Transport: tr, Timeout: isHealthyTimeout}
	} else {
		httpClient = &http.Client{Timeout: isHealthyTimeout}
	}

	return &Operator{
		cfg: cfg,
		//asgProvider:      asgProvider,
		snapshotProvider: snapshotProvider,
		httpClient:       httpClient,
		state:            "UNKNOWN",
		ticker:           time.NewTicker(loopInterval),
		shutdownChan:     shutdownChan,
	}
}

func (s *Operator) Run() {
	go s.webserver()

	for {
		if err := s.evaluate(); err != nil {
			log.WithError(err).Warn("could not evaluate cluster state")
			s.wait()
			continue
		}
		if err := s.execute(); err != nil {
			log.WithError(err).Warn("could not execute action")
		}
		s.wait()
	}
}

func (s *Operator) evaluate() error {
	// Fetch the auto-scaling group state.
	/*asgInstances, asgSelf, asgSize, err := s.asgProvider.AutoScalingGroupStatus()
	if err != nil {
		return fmt.Errorf("failed to sync auto-scaling group: %v", err)
	}*/

	// Create the etcd cluster client.
	client, err := etcd.NewClient(s.cfg.InitialMembersAddresses, etcd.SecurityConfig{
		CAFile:        s.cfg.Etcd.Ec.ClientTLSInfo.CAFile,
		CertFile:      s.cfg.Etcd.Ec.ClientTLSInfo.CertFile,
		KeyFile:       s.cfg.Etcd.Ec.ClientTLSInfo.KeyFile,
		CertAuth:      s.cfg.Etcd.Ec.ClientTLSInfo.ClientCertAuth,
		TrustedCAFile: s.cfg.Etcd.Ec.ClientTLSInfo.TrustedCAFile,
		AutoTLS:       s.cfg.Etcd.Ec.ClientAutoTLS,
	}, true)
	if err != nil {
		log.WithError(err).Warn("failed to create etcd cluster client")
	}

	// Output.
	if s.server == nil {
		s.server = etcd.NewServer(serverConfig(s.cfg, s.snapshotProvider), s.cfg.Etcd)
	}

	s.etcdRunning = s.server.IsRunning()
	s.etcdHealthy, s.isSeeder, s.states = fetchStatuses(s.httpClient, client, s.cfg.InitialMembersAddresses, s.cfg.CurrentMemberAddress)
	s.clusterSize = len(s.states)

	s.etcdClient = client
	return nil
}

func (s *Operator) execute() error {
	defer func() {
		if s.etcdClient != nil {
			s.etcdClient.Close()
		}
	}()

	switch {
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	case s.shutdown:
		log.Info("STATUS: Received SIGTERM -> Snapshot + Stop")
		s.state = "PENDING"

		s.server.Stop(s.etcdHealthy, true)
		os.Exit(0)
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	case s.etcdHealthy && !s.etcdRunning:
		log.Info("STATUS: Healthy + Not running -> Join")
		s.state = "PENDING"

		if err := s.server.Join(s.etcdClient); err != nil {
			log.WithError(err).Error("failed to join the cluster")
		}
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	case s.etcdHealthy && s.etcdRunning:
		log.Info("STATUS: Healthy + Running -> Standby")
		s.state = "OK"
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	case !s.etcdHealthy && s.etcdRunning && s.states["OK"] >= s.clusterSize/2+1:
		log.Info("STATUS: Unhealthy + Running -> Pending confirmation from other ECO instances")
		s.state = "PENDING"
	case !s.etcdHealthy && s.etcdRunning && s.states["OK"] < s.clusterSize/2+1:
		log.Info("STATUS: Unhealthy + Running + No quorum -> Snapshot + Stop")
		s.state = "PENDING"

		s.server.Stop(false, true)
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	case !s.etcdHealthy && !s.etcdRunning && (s.states["START"] != s.clusterSize || !s.isSeeder):
		if s.state != "START" {
			var err error
			if s.etcdSnapshot, err = s.server.SnapshotInfo(); err != nil && err != snapshot.ErrNoSnapshot {
				return err
			}
		}
		log.Info("STATUS: Unhealthy + Not running -> Ready to start + Pending all ready / seeder")
		s.state = "START"
		fmt.Println(s.etcdHealthy, "<>", s.etcdRunning, "<>", s.states, "<>", s.clusterSize, "<>", s.isSeeder)
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	case !s.etcdHealthy && !s.etcdRunning && s.states["START"] == s.clusterSize && s.isSeeder:
		log.Info("STATUS: Unhealthy + Not running + All ready + Seeder status -> Seeding cluster")
		s.state = "START"

		if err := s.server.Seed(s.etcdSnapshot); err != nil {
			log.WithError(err).Error("failed to seed the cluster")
		}
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	default:
		s.state = "UNKNOWN"
		return errors.New("no adequate action found")
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	}

	return nil
}

func (s *Operator) webserver() {
	http.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		st := status{State: s.state}
		if s.etcdSnapshot != nil {
			st.Revision = s.etcdSnapshot.Revision
		}
		b, err := json.Marshal(&st)
		if err != nil {
			log.WithError(err).Warn("failed to marshal status")
			return
		}
		if _, err := w.Write(b); err != nil {
			log.WithError(err).Warn("failed to write status")
		}
	})
	if s.cfg.Etcd.Ec.PeerAutoTLS || !s.cfg.Etcd.Ec.PeerTLSInfo.Empty() {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", webServerPort), s.cfg.Etcd.Ec.PeerTLSInfo.CertFile, s.cfg.Etcd.Ec.PeerTLSInfo.KeyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", webServerPort), nil))
	}
}

func (s *Operator) wait() {
	if s.etcdClient != nil {
		s.etcdClient.Close()
	}

	select {
	case <-s.ticker.C:
	case <-s.shutdownChan:
		s.shutdown = true
	}
}
