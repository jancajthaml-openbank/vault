package metrics

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/vault/utils"

	gom "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

type Metrics struct {
	promisesAccepted    gom.Counter
	commitsAccepted     gom.Counter
	rollbacksAccepted   gom.Counter
	createdAccounts     gom.Counter
	updatedSnapshots    gom.Meter
	snapshotCronLatency gom.Timer
}

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

func (gom *Metrics) TimeUpdateSaturatedSnapshots(f func()) {
	gom.snapshotCronLatency.Time(f)
}

func (gom *Metrics) SnapshotsUpdated(num int64) {
	gom.updatedSnapshots.Mark(num)
}

func (gom *Metrics) AccountCreated() {
	gom.createdAccounts.Inc(1)
}

func (gom *Metrics) PromiseAccepted() {
	gom.promisesAccepted.Inc(1)
}

func (gom *Metrics) CommitAccepted() {
	gom.commitsAccepted.Inc(1)
}

func (gom *Metrics) RollbackAccepted() {
	gom.rollbacksAccepted.Inc(1)
}

func (gom *Metrics) serialize() string {
	return "{\"CRON-SNAPSHOT-LATENCY-PCT-95\":" + strconv.FormatFloat(gom.snapshotCronLatency.Percentile(0.95), 'f', -1, 64) +
		",\"CRON-SNAPSHOTS-UPDATED\":" + strconv.FormatInt(gom.updatedSnapshots.Count(), 10) +
		",\"ACCOUNTS-CREATED\":" + strconv.FormatInt(gom.createdAccounts.Count(), 10) +
		",\"PROMISES-ACCEPTED\":" + strconv.FormatInt(gom.promisesAccepted.Count(), 10) +
		",\"COMMITS-ACCEPTED\":" + strconv.FormatInt(gom.commitsAccepted.Count(), 10) +
		",\"ROLLBACKS-ACCEPTED\":" + strconv.FormatInt(gom.rollbacksAccepted.Count(), 10) +
		"}"
}

func (gom *Metrics) persist(filename string) {
	tempFile := filename + "_temp"
	data := []byte(gom.serialize())
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

func PersistMetrics(wg *sync.WaitGroup, terminationChan chan struct{}, params utils.RunParams, data *Metrics) {
	defer wg.Done()

	if len(params.MetricsOutput) == 0 {
		log.Warnf("no metrics output defined, skipping metrics persistence")
		return
	}

	ticker := time.NewTicker(params.MetricsRefreshRate)
	defer ticker.Stop()

	log.Debugf("Updating metrics each %v into %v", params.MetricsRefreshRate, params.MetricsOutput)

	for {
		select {
		case <-ticker.C:
			data.persist(params.MetricsOutput)
		case <-terminationChan:
			data.persist(params.MetricsOutput)
			return
		}
	}
}
