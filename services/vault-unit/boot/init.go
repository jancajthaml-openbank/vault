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

package boot

import (
	"context"
	"os"

	"github.com/jancajthaml-openbank/vault-unit/actor"
	"github.com/jancajthaml-openbank/vault-unit/config"
	"github.com/jancajthaml-openbank/vault-unit/metrics"
	"github.com/jancajthaml-openbank/vault-unit/support/concurrent"
	"github.com/jancajthaml-openbank/vault-unit/support/logging"
)

// Program encapsulate initialized application
type Program struct {
	interrupt chan os.Signal
	cfg       config.Configuration
	daemons   []concurrent.Daemon
	cancel    context.CancelFunc
}

// NewProgram returns new program
func NewProgram() Program {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.LoadConfig()

	logging.SetupLogger(cfg.LogLevel)

	metricsDaemon := metrics.NewMetrics(
		ctx,
		cfg.MetricsOutput,
		cfg.Tenant,
		cfg.MetricsRefreshRate,
	)
	actorSystemDaemon := actor.NewActorSystem(
		ctx,
		cfg.Tenant,
		cfg.LakeHostname,
		cfg.SnapshotSaturationTreshold,
		cfg.RootStorage,
		metricsDaemon,
	)

	var daemons = make([]concurrent.Daemon, 0)
	daemons = append(daemons, metricsDaemon)
	daemons = append(daemons, actorSystemDaemon)

	return Program{
		interrupt: make(chan os.Signal, 1),
		cfg:       cfg,
		daemons:   daemons,
		cancel:    cancel,
	}
}
