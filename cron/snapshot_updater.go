package cron

import (
	"strconv"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	log "github.com/sirupsen/logrus"
)

type saturationCallback = func(utils.RunParams, *Metrics, interface{}, string, string)

// FIXME unit test coverage
// FIXME maximum events to params
func updateSaturated(params utils.RunParams, metrics *Metrics, callback saturationCallback) {
	log.Debugf("Scanning for saturated snapshots")

	accounts := getAccounts(params)
	var numberOfSnapshotsUpdated int64

	for _, name := range accounts {
		version := getVersion(params, name)
		if version == -1 {
			continue
		}
		if getEvents(params, name, version) >= params.JournalSaturation {
			updateAccount(params, metrics, name, version, version+1, callback)
			numberOfSnapshotsUpdated++
		}
	}
	metrics.SnapshotsUpdated(numberOfSnapshotsUpdated)
}

func updateAccount(params utils.RunParams, metrics *Metrics, name string, fromVersion, toVersion int, callback saturationCallback) {
	log.Debugf("Request %v to update snapshot version from %d to %d", name, fromVersion, toVersion)
	callback(params, metrics, model.Update{Version: fromVersion}, name, "snapshot_saturation_cron")
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
func SnapshotSaturationScan(wg *sync.WaitGroup, terminationChan chan struct{}, params utils.RunParams, metrics *Metrics, callback saturationCallback) {
	defer wg.Done()

	ticker := time.NewTicker(params.SnapshotScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics.TimeUpdateSaturatedSnapshots(func() {
				updateSaturated(params, metrics, callback)
			})
		case <-terminationChan:
			return
		}
	}
}
