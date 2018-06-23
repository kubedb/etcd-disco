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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/appscode/etcd-disco/pkg/etcd"
	"github.com/appscode/etcd-disco/pkg/providers/snapshot"
	log "github.com/sirupsen/logrus"
)

const (
	isHealthRetries  = 3
	isHealthyTimeout = 10 * time.Second
)

type status struct {
	instance string

	State    string `json:"state"`
	Revision int64  `json:"revision"`
}

func initSnapshotProvider(cfg snapshot.Config) snapshot.Provider {
	snapshotProvider, ok := snapshot.AsMap()[cfg.Provider]
	if !ok {
		log.Fatalf("unknown snapshot provider %q, available providers: %v", cfg.Provider, snapshot.AsList())
	}
	if err := snapshotProvider.Configure(cfg); err != nil {
		log.WithError(err).Fatal("failed to configure snapshot provider")
	}

	return snapshotProvider
}

func fetchStatuses(httpClient *http.Client, etcdClient *etcd.Client, Instances []string, self string) (bool, bool, map[string]int) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(1 + len(Instances))

	// Fetch etcd's healthiness.
	var etcdHealthy bool
	go func() {
		defer wg.Done()
		etcdHealthy = etcdClient != nil && etcdClient.IsHealthy(isHealthRetries, isHealthyTimeout)
	}()

	// Fetch ECO statuses.
	var ecoStatuses []*status
	for _, instance := range Instances {
		go func(instance string) {
			defer wg.Done()

			st, err := fetchStatus(httpClient, instance)
			if err != nil {
				log.WithError(err).Warnf("failed to query %s's instance", instance)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			ecoStatuses = append(ecoStatuses, st)
		}(instance)
	}
	wg.Wait()

	// Sort the ECO statuses so we can systematically find the identity of the seeder.
	sort.Slice(ecoStatuses, func(i, j int) bool {
		if ecoStatuses[i].Revision == ecoStatuses[j].Revision {
			return ecoStatuses[i].instance < ecoStatuses[j].instance
		}
		return ecoStatuses[i].Revision < ecoStatuses[j].Revision
	})

	// Count ECO statuses and determine if we are the seeder.
	ecoStates := make(map[string]int)
	for _, ecoStatus := range ecoStatuses {
		if _, ok := ecoStates[ecoStatus.State]; !ok {
			ecoStates[ecoStatus.State] = 0
		}
		ecoStates[ecoStatus.State]++
	}
	fmt.Println(ecoStatuses[len(ecoStatuses)-1].instance, self, "**********")

	return etcdHealthy, ecoStatuses[len(ecoStatuses)-1].instance == self, ecoStates
}

func fetchStatus(httpClient *http.Client, instance string) (*status, error) {
	scheme := "http"
	if httpClient.Transport != nil {
		scheme = "https"
	}
	resp, err := httpClient.Get(fmt.Sprintf("%s://%s:%d/status", scheme, instance, webServerPort))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var st status
	err = json.Unmarshal(b, &st)
	st.instance = instance
	fmt.Println(st, "<<>>>>>>>>>>>")
	return &st, err
}

func serverConfig(cfg Config, snapshotProvider snapshot.Provider) *etcd.ServerConfig {
	return &etcd.ServerConfig{
		UnhealthyMemberTTL: cfg.UnhealthyMemberTTL,
		SnapshotProvider:   snapshotProvider,
		SnapshotInterval:   cfg.Snapshot.Interval,
		SnapshotTTL:        cfg.Snapshot.TTL,
	}
}

func stringOverride(s, override string) string {
	if override != "" {
		return override
	}
	return s
}
