// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package metrics

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// TimeUpdateSaturatedSnapshots measures time of SaturatedSnapshots function run
func (metrics *Metrics) TimeUpdateSaturatedSnapshots(f func()) {
	metrics.snapshotCronLatency.Time(f)
}

// SnapshotsUpdated increments updated snapshots by given count
func (metrics *Metrics) SnapshotsUpdated(count int64) {
	metrics.updatedSnapshots.Mark(count)
}

// AccountCreated increments account created by one
func (metrics *Metrics) AccountCreated() {
	metrics.createdAccounts.Inc(1)
}

// PromiseAccepted increments accepted promises by one
func (metrics *Metrics) PromiseAccepted() {
	metrics.promisesAccepted.Inc(1)
}

// CommitAccepted increments accepted commits by one
func (metrics *Metrics) CommitAccepted() {
	metrics.commitsAccepted.Inc(1)
}

// RollbackAccepted increments accepted rollbacks by one
func (metrics *Metrics) RollbackAccepted() {
	metrics.rollbacksAccepted.Inc(1)
}

// Start handles everything needed to start metrics daemon
func (metrics Metrics) Start() {
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if err := metrics.Hydrate(); err != nil {
		log.Warn(err.Error())
	}
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		metrics.MarkDone()
		return
	}

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, metrics.output)

	go func() {
		for {
			select {
			case <-metrics.Done():
				metrics.Persist()
				metrics.MarkDone()
				return
			case <-ticker.C:
				metrics.Persist()
			}
		}
	}()

	<-metrics.IsDone
	log.Info("Stop metrics daemon")
}
