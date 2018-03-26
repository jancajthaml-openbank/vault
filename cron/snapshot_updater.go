package cron

import (
	"strconv"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/vault/actor"
	"github.com/jancajthaml-openbank/vault/metrics"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	log "github.com/sirupsen/logrus"
)

type saturationCallback = func(utils.RunParams, *metrics.Metrics, interface{}, string, actor.Coordinates)

// FIXME unit test coverage
// FIXME maximum events to params
func updateSaturated(params utils.RunParams, m *metrics.Metrics, callback saturationCallback) {
	accounts := getAccounts(params)
	var numberOfSnapshotsUpdated int64

	for _, name := range accounts {
		version := getVersion(params, name)
		if version == -1 {
			continue
		}
		if getEvents(params, name, version) >= params.Journal.JournalSaturation {
			updateAccount(params, m, name, version, version+1, callback)
			numberOfSnapshotsUpdated++
		}
	}
	m.SnapshotsUpdated(numberOfSnapshotsUpdated)
}

func updateAccount(params utils.RunParams, m *metrics.Metrics, name string, fromVersion, toVersion int, callback saturationCallback) {
	log.Debugf("Request %v to update snapshot version from %d to %d", name, fromVersion, toVersion)
	callback(params, m, model.Update{Version: fromVersion}, name, actor.Coordinates{"snapshot_saturation_cron", ""})
}

func getAccounts(params utils.RunParams) []string {
	return utils.ListDirectory(utils.AccountsPath(params), true)
}

func getVersion(params utils.RunParams, name string) int {
	versions := utils.ListDirectory(utils.SnapshotsPath(params, name), false)
	if len(versions) == 0 {
		return -1
	}

	version, err := strconv.Atoi(versions[0])
	if err != nil {
		return -1
	}

	return version
}

func getEvents(params utils.RunParams, name string, version int) int {
	return utils.CountNodes(utils.EventPath(params, name, version))
}

// SnapshotSaturationScan runs scan of accounts snapshots and events and orders
// accounts to update their snapshot if number of events in given version is
// larger than desiredd
func SnapshotSaturationScan(wg *sync.WaitGroup, terminationChan chan struct{}, params utils.RunParams, m *metrics.Metrics, callback saturationCallback) {
	defer wg.Done()

	ticker := time.NewTicker(params.Journal.SnapshotScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.TimeUpdateSaturatedSnapshots(func() {
				updateSaturated(params, m, callback)
			})
		case <-terminationChan:
			return
		}
	}
}
