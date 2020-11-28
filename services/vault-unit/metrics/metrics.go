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
	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics holds metrics counters
type Metrics struct {
	storage             localfs.Storage
	tenant              string
	promisesAccepted    metrics.Counter
	commitsAccepted     metrics.Counter
	rollbacksAccepted   metrics.Counter
	createdAccounts     metrics.Counter
	updatedSnapshots    metrics.Meter
	snapshotCronLatency metrics.Timer
}

// NewMetrics returns blank metrics holder
func NewMetrics(output string, tenant string) *Metrics {
	storage, err := localfs.NewPlaintextStorage(output)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &Metrics{
		storage:             storage,
		tenant:              tenant,
		promisesAccepted:    metrics.NewCounter(),
		commitsAccepted:     metrics.NewCounter(),
		rollbacksAccepted:   metrics.NewCounter(),
		createdAccounts:     metrics.NewCounter(),
		updatedSnapshots:    metrics.NewMeter(),
		snapshotCronLatency: metrics.NewTimer(),
	}
}

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

func (metrics *Metrics) Setup() error {
	return nil
}

func (metrics *Metrics) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}

func (metrics *Metrics) Cancel() {
}

// Work represents metrics worker work
func (metrics *Metrics) Work() {
	metrics.Persist()
}
