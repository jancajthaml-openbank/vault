// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	localfs "github.com/jancajthaml-openbank/local-fs"

	"github.com/jancajthaml-openbank/vault-unit/actor"
	"github.com/jancajthaml-openbank/vault-unit/config"
	"github.com/jancajthaml-openbank/vault-unit/metrics"
	"github.com/jancajthaml-openbank/vault-unit/persistence"
	"github.com/jancajthaml-openbank/vault-unit/utils"
)

// Application encapsulate initialized application
type Application struct {
	cfg             config.Configuration
	interrupt       chan os.Signal
	metrics         metrics.Metrics
	actorSystem     actor.ActorSystem
	snapshotUpdater persistence.SnapshotUpdater
	cancel          context.CancelFunc
}

// Initialize application
func Initialize() Application {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.GetConfig()

	utils.SetupLogger(cfg.LogLevel)

	storage := localfs.NewStorage(cfg.RootStorage)
	metricsDaemon := metrics.NewMetrics(ctx, cfg.Tenant, cfg.MetricsOutput, cfg.MetricsRefreshRate)

	actorSystemDaemon := actor.NewActorSystem(ctx, cfg.Tenant, cfg.LakeHostname, &metricsDaemon, &storage)
	snapshotUpdaterDaemon := persistence.NewSnapshotUpdater(ctx, cfg.JournalSaturation, cfg.SnapshotScanInterval, &metricsDaemon, &storage, actor.ProcessLocalMessage(&actorSystemDaemon))

	return Application{
		cfg:             cfg,
		interrupt:       make(chan os.Signal, 1),
		metrics:         metricsDaemon,
		actorSystem:     actorSystemDaemon,
		snapshotUpdater: snapshotUpdaterDaemon,
		cancel:          cancel,
	}
}
