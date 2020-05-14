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

package persistence

import (
	"context"
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/vault-unit/metrics"
	"github.com/jancajthaml-openbank/vault-unit/model"
	"github.com/jancajthaml-openbank/vault-unit/utils"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// SnapshotUpdater represents journal saturation update subroutine
type SnapshotUpdater struct {
	utils.DaemonSupport
	callback            func(msg interface{}, to system.Coordinates, from system.Coordinates)
	metrics             *metrics.Metrics
	storage             *localfs.Storage
	scanInterval        time.Duration
	saturationThreshold int
}

// NewSnapshotUpdater returns snapshot updater fascade
func NewSnapshotUpdater(ctx context.Context, saturation int, scanInterval time.Duration, metrics *metrics.Metrics, storage *localfs.Storage, callback func(msg interface{}, to system.Coordinates, from system.Coordinates)) SnapshotUpdater {
	return SnapshotUpdater{
		DaemonSupport:       utils.NewDaemonSupport(ctx, "snapshot-updater"),
		callback:            callback,
		metrics:             metrics,
		storage:             storage,
		scanInterval:        scanInterval,
		saturationThreshold: saturation,
	}
}

// FIXME unit test coverage
// FIXME maximum events to params
func (updater SnapshotUpdater) updateSaturated() {
	accounts := updater.getAccounts()
	var numberOfSnapshotsUpdated int64

	for _, name := range accounts {
		version := updater.getVersion(name)
		if version == -1 {
			continue
		}
		if updater.getEvents(name, version) >= updater.saturationThreshold {
			log.Debugf("Request %v to update snapshot version from %d to %d", name, version, version+1)
			msg := model.Update{Version: version}
			to := system.Coordinates{Name: name}
			from := system.Coordinates{Name: "snapshot_saturation_cron"}
			updater.callback(msg, to, from)

			numberOfSnapshotsUpdated++
		}
	}
	updater.metrics.SnapshotsUpdated(numberOfSnapshotsUpdated)
}

func (updater SnapshotUpdater) getAccounts() []string {
	result, err := updater.storage.ListDirectory(utils.RootPath(), true)
	if err != nil {
		return nil
	}
	return result
}

func (updater SnapshotUpdater) getVersion(name string) int {
	result, err := updater.storage.ListDirectory(utils.SnapshotsPath(name), false)
	if err != nil || len(result) == 0 {
		return -1
	}

	version, err := strconv.Atoi(result[0])
	if err != nil {
		return -1
	}

	return version
}

func (updater SnapshotUpdater) getEvents(name string, version int) int {
	result, err := updater.storage.CountFiles(utils.EventPath(name, version))
	if err != nil {
		return -1
	}
	return result
}

// Start handles everything needed to start snapshot updater daemon it runs scan
// of accounts snapshots and events and orders accounts to update their snapshot
// if number of events in given version is larger than threshold
func (updater SnapshotUpdater) Start() {
	ticker := time.NewTicker(updater.scanInterval)
	defer ticker.Stop()

	updater.MarkReady()

	select {
	case <-updater.CanStart:
		break
	case <-updater.Done():
		updater.MarkDone()
		return
	}

	log.Infof("Start snapshot-updater daemon, scan each %v and update journals with at least %d events", updater.scanInterval, updater.saturationThreshold)

	go func() {
		for {
			select {
			case <-updater.Done():
				updater.MarkDone()
				return
			case <-ticker.C:
				//monitor.CheckMemoryAllocation()
				updater.metrics.TimeUpdateSaturatedSnapshots(func() {
					updater.updateSaturated()
				})
			}
		}
	}()

	<-updater.IsDone
	log.Info("Stop snapshot-updater daemon")
}
