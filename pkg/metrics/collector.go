package metrics

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/vault/pkg/utils"

	gom "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Snapshot holds metrics snapshot status
type Snapshot struct {
	SnapshotCronLatency float64 `json:"snapshotCronLatency"`
	UpdatedSnapshots    int64   `json:"updatedSnapshots"`
	CreatedAccounts     int64   `json:"createdAccounts"`
	PromisesAccepted    int64   `json:"promisesAccepted"`
	CommitsAccepted     int64   `json:"commitsAccepted"`
	RollbacksAccepted   int64   `json:"rollbacksAccepted"`
}

// Metrics holds metrics counters
type Metrics struct {
	promisesAccepted    gom.Counter
	commitsAccepted     gom.Counter
	rollbacksAccepted   gom.Counter
	createdAccounts     gom.Counter
	updatedSnapshots    gom.Meter
	snapshotCronLatency gom.Timer
}

// NewMetrics returns blank metrics holder
func NewMetrics() *Metrics {
	return &Metrics{
		promisesAccepted:    gom.NewCounter(),
		commitsAccepted:     gom.NewCounter(),
		rollbacksAccepted:   gom.NewCounter(),
		createdAccounts:     gom.NewCounter(),
		updatedSnapshots:    gom.NewMeter(),
		snapshotCronLatency: gom.NewTimer(),
	}
}

// NewSnapshot returns metrics snapshot
func NewSnapshot(entity *Metrics) Snapshot {
	if entity == nil {
		return Snapshot{}
	}

	return Snapshot{
		SnapshotCronLatency: entity.snapshotCronLatency.Percentile(0.95),
		UpdatedSnapshots:    entity.updatedSnapshots.Count(),
		CreatedAccounts:     entity.createdAccounts.Count(),
		PromisesAccepted:    entity.promisesAccepted.Count(),
		CommitsAccepted:     entity.commitsAccepted.Count(),
		RollbacksAccepted:   entity.rollbacksAccepted.Count(),
	}
}

// TimeUpdateSaturatedSnapshots measures time of SaturatedSnapshots function run
func (entity *Metrics) TimeUpdateSaturatedSnapshots(f func()) {
	entity.snapshotCronLatency.Time(f)
}

// SnapshotsUpdated increments updated snapshots by given count
func (entity *Metrics) SnapshotsUpdated(count int64) {
	entity.updatedSnapshots.Mark(count)
}

// AccountCreated increments account created by one
func (entity *Metrics) AccountCreated() {
	entity.createdAccounts.Inc(1)
}

// PromiseAccepted increments accepted promises by one
func (entity *Metrics) PromiseAccepted() {
	entity.promisesAccepted.Inc(1)
}

// CommitAccepted increments accepted commits by one
func (entity *Metrics) CommitAccepted() {
	entity.commitsAccepted.Inc(1)
}

// RollbackAccepted increments accepted rollbacks by one
func (entity *Metrics) RollbackAccepted() {
	entity.rollbacksAccepted.Inc(1)
}

func (entity *Metrics) persist(filename string) {
	tempFile := filename + "_temp"

	data, err := json.Marshal(NewSnapshot(entity))
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

	log.Debugf("metrics updated at %s", filename)
	return
}

// PersistMetrics stores metrics holded in memory periodically to disk
func PersistMetrics(wg *sync.WaitGroup, terminationChan chan struct{}, params utils.RunParams, data *Metrics) {
	defer wg.Done()

	if params.Metrics.Output == "" {
		log.Warnf("no metrics output defined, skipping metrics persistence")
		return
	}

	output := params.Metrics.Output + "." + params.Setup.Tenant

	ticker := time.NewTicker(params.Metrics.RefreshRate)
	defer ticker.Stop()

	log.Debugf("Updating metrics each %v into %v", params.Metrics.RefreshRate, output)

	for {
		select {
		case <-ticker.C:
			data.persist(output)
		case <-terminationChan:
			data.persist(output)
			return
		}
	}
}
