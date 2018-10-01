// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package daemon

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/jancajthaml-openbank/vault/config"

	gom "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Snapshot holds metrics snapshot status
type Snapshot struct {
	SnapshotCronLatency float64 `json:"snapshotCronLatency"`
	UpdatedSnapshots    int64   `json:"updatedSnapshots"`
	CreatedAccounts     int64   `json:"createdAccounts"`
	PromisesAccepted    int64   `json:"promisesAccepted"`
	CommitsAccepted     int64   `json:"commitsAccepted"`
	RollbacksAccepted   int64   `json:"rollbacksAccepted"`
}

// Metrics holds metrics counters
type Metrics struct {
	Support
	output              string
	tenant              string
	refreshRate         time.Duration
	promisesAccepted    gom.Counter
	commitsAccepted     gom.Counter
	rollbacksAccepted   gom.Counter
	createdAccounts     gom.Counter
	updatedSnapshots    gom.Meter
	snapshotCronLatency gom.Timer
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, cfg config.Configuration) Metrics {
	return Metrics{
		Support:             NewDaemonSupport(ctx),
		output:              cfg.MetricsOutput,
		tenant:              cfg.Tenant,
		refreshRate:         cfg.MetricsRefreshRate,
		promisesAccepted:    gom.NewCounter(),
		commitsAccepted:     gom.NewCounter(),
		rollbacksAccepted:   gom.NewCounter(),
		createdAccounts:     gom.NewCounter(),
		updatedSnapshots:    gom.NewMeter(),
		snapshotCronLatency: gom.NewTimer(),
	}
}

// NewSnapshot returns metrics snapshot
func NewSnapshot(gom Metrics) Snapshot {
	return Snapshot{
		SnapshotCronLatency: gom.snapshotCronLatency.Percentile(0.95),
		UpdatedSnapshots:    gom.updatedSnapshots.Count(),
		CreatedAccounts:     gom.createdAccounts.Count(),
		PromisesAccepted:    gom.promisesAccepted.Count(),
		CommitsAccepted:     gom.commitsAccepted.Count(),
		RollbacksAccepted:   gom.rollbacksAccepted.Count(),
	}
}

// TimeUpdateSaturatedSnapshots measures time of SaturatedSnapshots function run
func (gom Metrics) TimeUpdateSaturatedSnapshots(f func()) {
	gom.snapshotCronLatency.Time(f)
}

// SnapshotsUpdated increments updated snapshots by given count
func (gom Metrics) SnapshotsUpdated(count int64) {
	gom.updatedSnapshots.Mark(count)
}

// AccountCreated increments account created by one
func (gom Metrics) AccountCreated() {
	gom.createdAccounts.Inc(1)
}

// PromiseAccepted increments accepted promises by one
func (gom Metrics) PromiseAccepted() {
	gom.promisesAccepted.Inc(1)
}

// CommitAccepted increments accepted commits by one
func (gom Metrics) CommitAccepted() {
	gom.commitsAccepted.Inc(1)
}

// RollbackAccepted increments accepted rollbacks by one
func (gom Metrics) RollbackAccepted() {
	gom.rollbacksAccepted.Inc(1)
}

func (gom Metrics) persist(filename string) {
	tempFile := filename + "_temp"

	data, err := json.Marshal(NewSnapshot(gom))
	if err != nil {
		log.Warnf("unable to create serialize metrics with error: %v", err)
		return
	}
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Warnf("unable to create file with error: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		log.Warnf("unable to write file with error: %v", err)
		return
	}

	if err := os.Rename(tempFile, filename); err != nil {
		log.Warnf("unable to move file with error: %v", err)
		return
	}

	return
}

// Start handles everything needed to start metrics daemon
func (gom Metrics) Start() {
	defer gom.MarkDone()

	if gom.output == "" {
		log.Warnf("no metrics output defined, skipping metrics persistence")
		gom.MarkReady()
		return
	}

	ticker := time.NewTicker(gom.refreshRate)
	defer ticker.Stop()

	log.Infof("Start metrics daemon, update each %v into %v/%v", gom.refreshRate, gom.output, gom.tenant)

	gom.MarkReady()

	for {
		select {
		case <-gom.Done():
			log.Info("Stopping metrics daemon")
			gom.persist(gom.output + "/" + gom.tenant)
			log.Info("Stop metrics daemon")
			return
		case <-ticker.C:
			gom.persist(gom.output + "/" + gom.tenant)
		}
	}
}
