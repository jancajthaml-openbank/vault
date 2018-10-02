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
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/vault/actor"
	"github.com/jancajthaml-openbank/vault/config"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	log "github.com/sirupsen/logrus"
)

// SnapshotUpdater represents journal saturation update subroutine
type SnapshotUpdater struct {
	Support
	callback            func(msg interface{}, receiver string, sender actor.Coordinates)
	metrics             *Metrics
	rootStorage         string
	scanInterval        time.Duration
	saturationThreshold int
}

// NewSnapshotUpdater returns snapshot updater fascade
func NewSnapshotUpdater(ctx context.Context, cfg config.Configuration, metrics *Metrics, callback func(msg interface{}, receiver string, sender actor.Coordinates)) SnapshotUpdater {
	return SnapshotUpdater{
		Support:             NewDaemonSupport(ctx),
		callback:            callback,
		metrics:             metrics,
		rootStorage:         cfg.RootStorage,
		scanInterval:        cfg.SnapshotScanInterval,
		saturationThreshold: cfg.JournalSaturation,
	}
}

// FIXME unit test coverage
// FIXME maximum events to params
func (su SnapshotUpdater) updateSaturated() {
	accounts := su.getAccounts()
	var numberOfSnapshotsUpdated int64

	for _, name := range accounts {
		version := su.getVersion(name)
		if version == -1 {
			continue
		}
		if su.getEvents(name, version) >= su.saturationThreshold {
			su.updateAccount(name, version, version+1)
			numberOfSnapshotsUpdated++
		}
	}
	su.metrics.SnapshotsUpdated(numberOfSnapshotsUpdated)
}

func (su SnapshotUpdater) updateAccount(name string, fromVersion, toVersion int) {
	log.Debugf("Request %v to update snapshot version from %d to %d", name, fromVersion, toVersion)
	msg := model.Update{Version: fromVersion}
	coordinates := actor.Coordinates{Name: "snapshot_saturation_cron"}
	su.callback(msg, name, coordinates)
}

func (su SnapshotUpdater) getAccounts() []string {
	return utils.ListDirectory(utils.AccountsPath(su.rootStorage), true)
}

func (su SnapshotUpdater) getVersion(name string) int {
	versions := utils.ListDirectory(utils.SnapshotsPath(su.rootStorage, name), false)
	if len(versions) == 0 {
		return -1
	}

	version, err := strconv.Atoi(versions[0])
	if err != nil {
		return -1
	}

	return version
}

func (su SnapshotUpdater) getEvents(name string, version int) int {
	return utils.CountFiles(utils.EventPath(su.rootStorage, name, version))
}

// Start handles everything needed to start snapshot updater daemon it runs scan
// of accounts snapshots and events and orders accounts to update their snapshot
// if number of events in given version is larger than threshold
func (su SnapshotUpdater) Start() {
	defer su.MarkDone()

	ticker := time.NewTicker(su.scanInterval)
	defer ticker.Stop()

	log.Infof("Start snapshot updater daemon, scan each %v and update journals with at least %d events", su.scanInterval, su.saturationThreshold)

	su.MarkReady()

	for {
		select {
		case <-su.Done():
			log.Info("Stop snapshot updater daemon")
			return
		case <-ticker.C:
			su.metrics.TimeUpdateSaturatedSnapshots(func() {
				su.updateSaturated()
			})
		}
	}
}
