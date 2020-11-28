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
	storage              localfs.Storage
	getAccountLatency    metrics.Timer
	createAccountLatency metrics.Timer
}

// NewMetrics returns blank metrics holder
func NewMetrics(output string) *Metrics {
	storage, err := localfs.NewPlaintextStorage(output)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &Metrics{
		storage:              storage,
		getAccountLatency:    metrics.NewTimer(),
		createAccountLatency: metrics.NewTimer(),
	}
}

// TimeGetAccount measure execution of GetAccount
func (metrics *Metrics) TimeGetAccount(f func()) {
	if metrics == nil {
		return
	}
	metrics.getAccountLatency.Time(f)
}

// TimeCreateAccount measure execution of CreateAccount
func (metrics *Metrics) TimeCreateAccount(f func()) {
	if metrics == nil {
		return
	}
	metrics.createAccountLatency.Time(f)
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
