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
	"github.com/jancajthaml-openbank/vault-unit/logging"
	"github.com/jancajthaml-openbank/vault-unit/metrics"
	"github.com/jancajthaml-openbank/vault-unit/utils"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// Program encapsulate initialized application
type Program struct {
	interrupt chan os.Signal
	cfg       config.Configuration
	daemons   []utils.Daemon
	cancel    context.CancelFunc
}

// Initialize application
func Initialize() Program {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.GetConfig()

	logging.SetupLogger(cfg.LogLevel)

	storage := localfs.NewPlaintextStorage(
		cfg.RootStorage,
	)
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
		&metricsDaemon,
		&storage,
	)
	snapshotUpdaterDaemon := actor.NewSnapshotUpdater(
		ctx,
		cfg.JournalSaturation,
		cfg.SnapshotScanInterval,
		&metricsDaemon,
		&storage,
		func(account string, version int64) {
			ref, err := actor.NewAccountActor(&actorSystemDaemon, account)
			if err != nil {
				return
			}
			ref.Tell(
				actor.RequestUpdate{
					Version: version,
				},
				system.Coordinates{
					Region: actorSystemDaemon.Name,
					Name:   account,
				},
				system.Coordinates{
					Region: actorSystemDaemon.Name,
					Name:   "snapshot_updater_cron",
				},
			)
		},
	)

	var daemons = make([]utils.Daemon, 0)
	daemons = append(daemons, metricsDaemon)
	daemons = append(daemons, actorSystemDaemon)
	daemons = append(daemons, snapshotUpdaterDaemon)

	return Program{
		interrupt: make(chan os.Signal, 1),
		cfg:       cfg,
		daemons:   daemons,
		cancel:    cancel,
	}
}
