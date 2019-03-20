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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jancajthaml-openbank/vault-unit/config"
	"github.com/jancajthaml-openbank/vault-unit/utils"

	metrics "github.com/rcrowley/go-metrics"
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

// Metrics represents metrics subroutine
type Metrics struct {
	Support
	output              string
	tenant              string
	refreshRate         time.Duration
	promisesAccepted    metrics.Counter
	commitsAccepted     metrics.Counter
	rollbacksAccepted   metrics.Counter
	createdAccounts     metrics.Counter
	updatedSnapshots    metrics.Meter
	snapshotCronLatency metrics.Timer
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, cfg config.Configuration) Metrics {
	return Metrics{
		Support:             NewDaemonSupport(ctx),
		output:              cfg.MetricsOutput,
		tenant:              cfg.Tenant,
		refreshRate:         cfg.MetricsRefreshRate,
		promisesAccepted:    metrics.NewCounter(),
		commitsAccepted:     metrics.NewCounter(),
		rollbacksAccepted:   metrics.NewCounter(),
		createdAccounts:     metrics.NewCounter(),
		updatedSnapshots:    metrics.NewMeter(),
		snapshotCronLatency: metrics.NewTimer(),
	}
}

// NewSnapshot returns metrics snapshot
func NewSnapshot(metrics Metrics) Snapshot {
	return Snapshot{
		SnapshotCronLatency: metrics.snapshotCronLatency.Percentile(0.95),
		UpdatedSnapshots:    metrics.updatedSnapshots.Count(),
		CreatedAccounts:     metrics.createdAccounts.Count(),
		PromisesAccepted:    metrics.promisesAccepted.Count(),
		CommitsAccepted:     metrics.commitsAccepted.Count(),
		RollbacksAccepted:   metrics.rollbacksAccepted.Count(),
	}
}

// TimeUpdateSaturatedSnapshots measures time of SaturatedSnapshots function run
func (metrics Metrics) TimeUpdateSaturatedSnapshots(f func()) {
	metrics.snapshotCronLatency.Time(f)
}

// SnapshotsUpdated increments updated snapshots by given count
func (metrics Metrics) SnapshotsUpdated(count int64) {
	metrics.updatedSnapshots.Mark(count)
}

// AccountCreated increments account created by one
func (metrics Metrics) AccountCreated() {
	metrics.createdAccounts.Inc(1)
}

// PromiseAccepted increments accepted promises by one
func (metrics Metrics) PromiseAccepted() {
	metrics.promisesAccepted.Inc(1)
}

// CommitAccepted increments accepted commits by one
func (metrics Metrics) CommitAccepted() {
	metrics.commitsAccepted.Inc(1)
}

// RollbackAccepted increments accepted rollbacks by one
func (metrics Metrics) RollbackAccepted() {
	metrics.rollbacksAccepted.Inc(1)
}

func (metrics Metrics) persist(filename string) {
	tempFile := filename + "_temp"

	data, err := utils.JSON.Marshal(NewSnapshot(metrics))
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

func getFilename(path string, tenant string) string {
	if tenant == "" {
		return path
	}

	dirname := filepath.Dir(path)
	ext := filepath.Ext(path)
	filename := filepath.Base(path)
	filename = filename[:len(filename)-len(ext)]

	return dirname + "/" + filename + "." + tenant + ext
}

// WaitReady wait for metrics to be ready
func (metrics Metrics) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-metrics.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start metrics daemon
func (metrics Metrics) Start() {
	defer metrics.MarkDone()

	if metrics.output == "" {
		log.Warnf("no metrics output defined, skipping metrics persistence")
		metrics.MarkReady()
		return
	}

	output := getFilename(metrics.output, metrics.tenant)
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	metrics.MarkReady()

	select {
	case <-metrics.canStart:
		break
	case <-metrics.Done():
		return
	}

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, output)

	for {
		select {
		case <-metrics.Done():
			log.Info("Stopping metrics daemon")
			metrics.persist(output)
			log.Info("Stop metrics daemon")
			return
		case <-ticker.C:
			metrics.persist(output)
		}
	}
}
