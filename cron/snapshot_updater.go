package cron

import (
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	log "github.com/sirupsen/logrus"
)

type saturationCallback = func(utils.RunParams, interface{}, string, string)

// FIXME maximum events to params
func updateSaturated(params utils.RunParams, callback saturationCallback) {
	log.Debugf("Scanning for saturated snapshots")

	accounts := getAccounts(params)
	for _, name := range accounts {
		version := getVersion(params, name)
		if version == -1 {
			continue
		}
		if getEvents(params, name, version) > params.JournalSaturation {
			updateAccount(params, name, version, version+1, callback)
		}
	}
}

// SnapshotSaturationScan runs scan of accounts snapshots and events and orders
// accounts to update their snapshot if number of events in given version is
// larger than desiredd
func SnapshotSaturationScan(params utils.RunParams, callback saturationCallback) {
	// FIXME to params
	duration := 10 * time.Second

	ticker := time.NewTicker(duration)
	for range ticker.C {
		updateSaturated(params, callback)
	}
}

func updateAccount(params utils.RunParams, name string, fromVersion, toVersion int, callback saturationCallback) {
	log.Debugf("Request %v to update snapshot version from %d to %d", name, fromVersion, toVersion)
	callback(params, model.Update{Version: fromVersion}, name, "snapshot_saturation_cron")
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
