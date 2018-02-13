package cron

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/vault/utils"

	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

type Metrics struct {
	promisesAccepted    metrics.Counter
	commitsAccepted     metrics.Counter
	rollbacksAccepted   metrics.Counter
	createdAccounts     metrics.Counter
	updatedSnapshots    metrics.Meter
	snapshotCronLatency metrics.Timer
}

func NewMetrics() *Metrics {
	return &Metrics{
		promisesAccepted:    metrics.NewCounter(),
		commitsAccepted:     metrics.NewCounter(),
		rollbacksAccepted:   metrics.NewCounter(),
		createdAccounts:     metrics.NewCounter(),
		updatedSnapshots:    metrics.NewMeter(),
		snapshotCronLatency: metrics.NewTimer(),
	}
}

func (metrics *Metrics) TimeUpdateSaturatedSnapshots(f func()) {
	metrics.snapshotCronLatency.Time(f)
}

func (metrics *Metrics) SnapshotsUpdated(num int64) {
	metrics.updatedSnapshots.Mark(num)
}

func (metrics *Metrics) AccountCreated() {
	metrics.createdAccounts.Inc(1)
}

func (metrics *Metrics) PromiseAccepted() {
	metrics.promisesAccepted.Inc(1)
}

func (metrics *Metrics) CommitAccepted() {
	metrics.commitsAccepted.Inc(1)
}

func (metrics *Metrics) RollbackAccepted() {
	metrics.rollbacksAccepted.Inc(1)
}

func (metrics *Metrics) serialize() string {
	return "{\"CRON-SNAPSHOT-LATENCY-PCT-95\":" + strconv.FormatFloat(metrics.snapshotCronLatency.Percentile(0.95), 'f', -1, 64) +
		",\"CRON-SNAPSHOTS-UPDATED\":" + strconv.FormatInt(metrics.updatedSnapshots.Count(), 10) +
		",\"ACCOUNTS-CREATED\":" + strconv.FormatInt(metrics.createdAccounts.Count(), 10) +
		",\"PROMISES-ACCEPTED\":" + strconv.FormatInt(metrics.promisesAccepted.Count(), 10) +
		",\"COMMITS-ACCEPTED\":" + strconv.FormatInt(metrics.commitsAccepted.Count(), 10) +
		",\"ROLLBACKS-ACCEPTED\":" + strconv.FormatInt(metrics.rollbacksAccepted.Count(), 10) +
		"}"
}

func (metrics *Metrics) persist(filename string) {
	tempFile := filename + "_temp"
	data := []byte(metrics.serialize())
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

func (metrics *Metrics) dump() {
	// FIXME shortened format
	log.Debugf("Metrics | %s", metrics.serialize())
	return
}

func PersistMetrics(wg *sync.WaitGroup, terminationChan chan struct{}, params utils.RunParams, metrics *Metrics) {
	defer wg.Done()

	ticker := time.NewTicker(params.MetricsRefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(params.MetricsOutput) == 0 {
				metrics.dump()
			} else {
				metrics.persist(params.MetricsOutput)
			}
		case <-terminationChan:
			return
		}
	}
}
