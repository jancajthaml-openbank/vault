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

	"github.com/jancajthaml-openbank/vault-rest/config"
	"github.com/jancajthaml-openbank/vault-rest/utils"

	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Metrics represents metrics subroutine
type Metrics struct {
	Support
	output               string
	refreshRate          time.Duration
	getAccountLatency    metrics.Timer
	createAccountLatency metrics.Timer
	createdAccounts      metrics.Counter
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, cfg config.Configuration) Metrics {
	return Metrics{
		Support:              NewDaemonSupport(ctx),
		output:               cfg.MetricsOutput,
		refreshRate:          cfg.MetricsRefreshRate,
		getAccountLatency:    metrics.NewTimer(),
		createAccountLatency: metrics.NewTimer(),
		createdAccounts:      metrics.NewCounter(),
	}
}

// Snapshot holds metrics snapshot status
type Snapshot struct {
	GetAccountLatency    float64 `json:"getAccountLatency"`
	CreateAccountLatency float64 `json:"createAccountLatency"`
	CreatedAccounts      int64   `json:"createdAccounts"`
}

// NewSnapshot returns metrics snapshot
func NewSnapshot(metrics Metrics) Snapshot {
	return Snapshot{
		GetAccountLatency:    metrics.getAccountLatency.Percentile(0.95),
		CreateAccountLatency: metrics.createAccountLatency.Percentile(0.95),
		CreatedAccounts:      metrics.createdAccounts.Count(),
	}
}

// AccountCreated increments account created by one
func (metrics Metrics) AccountCreated() {
	metrics.createdAccounts.Inc(1)
}

// TimeGetAccount measure execution of GetAccount
func (metrics Metrics) TimeGetAccount(f func()) {
	metrics.getAccountLatency.Time(f)
}

// TimeCreateAccount measure execution of CreateAccount
func (metrics Metrics) TimeCreateAccount(f func()) {
	metrics.createAccountLatency.Time(f)
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

func getFilename(path string) string {
	dirname := filepath.Dir(path)
	ext := filepath.Ext(path)
	filename := filepath.Base(path)
	filename = filename[:len(filename)-len(ext)]

	return dirname + "/" + filename + ext
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

	output := getFilename(metrics.output)
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
