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
	"sync/atomic"

	"github.com/DataDog/datadog-go/statsd"
)

// Metrics provides helper function for metrics
type Metrics interface {
	SnapshotsUpdated(count int64)
	AccountCreated()
	PromiseAccepted()
	CommitAccepted()
	RollbackAccepted()
}

// StatsdMetrics provides metrics helper with statsd client
type StatsdMetrics struct {
	client            *statsd.Client
	tenant            string
	promisesAccepted  int64
	commitsAccepted   int64
	rollbacksAccepted int64
	createdAccounts   int64
	updatedSnapshots  int64
}

// NewMetrics helper function for metrics with statsd client
func NewMetrics(tenant string, endpoint string) *StatsdMetrics {
	client, err := statsd.New(endpoint, statsd.WithClientSideAggregation(), statsd.WithoutTelemetry())
	if err != nil {
		log.Error().Msgf("Failed to ensure statsd client %+v", err)
		return nil
	}
	return &StatsdMetrics{
		client:            client,
		tenant:            tenant,
		promisesAccepted:  int64(0),
		commitsAccepted:   int64(0),
		rollbacksAccepted: int64(0),
		createdAccounts:   int64(0),
		updatedSnapshots:  int64(0),
	}
}

// SnapshotsUpdated increments updated snapshots by given count
func (instance *StatsdMetrics) SnapshotsUpdated(count int64) {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.updatedSnapshots), count)
}

// AccountCreated increments account created by one
func (instance *StatsdMetrics) AccountCreated() {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.createdAccounts), 1)
}

// PromiseAccepted increments accepted promises by one
func (instance *StatsdMetrics) PromiseAccepted() {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.promisesAccepted), 1)
}

// CommitAccepted increments accepted commits by one
func (instance *StatsdMetrics) CommitAccepted() {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.commitsAccepted), 1)
}

// RollbackAccepted increments accepted rollbacks by one
func (instance *StatsdMetrics) RollbackAccepted() {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.rollbacksAccepted), 1)
}

// Setup does nothing
func (*StatsdMetrics) Setup() error {
	return nil
}

// Done returns always finished
func (*StatsdMetrics) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}

// Cancel triggers work once
func (instance *StatsdMetrics) Cancel() {
	instance.Work()
}

// Work represents metrics worker work
func (instance *StatsdMetrics) Work() {
	if instance == nil {
		return
	}

	accountCreated := instance.createdAccounts
	accountUpdated := instance.updatedSnapshots
	promisesAccepted := instance.promisesAccepted
	commitsAccepted := instance.commitsAccepted
	rollbacksAccepted := instance.rollbacksAccepted

	atomic.AddInt64(&(instance.createdAccounts), -accountCreated)
	atomic.AddInt64(&(instance.updatedSnapshots), -accountUpdated)
	atomic.AddInt64(&(instance.promisesAccepted), -promisesAccepted)
	atomic.AddInt64(&(instance.commitsAccepted), -commitsAccepted)
	atomic.AddInt64(&(instance.rollbacksAccepted), -rollbacksAccepted)

	tags := []string{"tenant:" + instance.tenant}

	instance.client.Count("openbank.vault.account.created", accountCreated, tags, 1)
	instance.client.Count("openbank.vault.account.updated", accountUpdated, tags, 1)
	instance.client.Count("openbank.vault.promise.accepted", promisesAccepted, tags, 1)
	instance.client.Count("openbank.vault.promise.committed", commitsAccepted, tags, 1)
	instance.client.Count("openbank.vault.promise.rollbacked", rollbacksAccepted, tags, 1)
}
