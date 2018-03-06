package metrics

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsPersist(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.RemoveAll(tmpDir)

	m := NewMetrics()

	t.Log("Proper values are updated")
	{
		require.Equal(t, int64(0), m.updatedSnapshots.Count())
		m.SnapshotsUpdated(2)
		assert.Equal(t, int64(2), m.updatedSnapshots.Count())

		require.Equal(t, int64(0), m.createdAccounts.Count())
		m.AccountCreated()
		assert.Equal(t, int64(1), m.createdAccounts.Count())

		require.Equal(t, int64(0), m.promisesAccepted.Count())
		m.PromiseAccepted()
		assert.Equal(t, int64(1), m.promisesAccepted.Count())

		require.Equal(t, int64(0), m.commitsAccepted.Count())
		m.CommitAccepted()
		assert.Equal(t, int64(1), m.commitsAccepted.Count())

		require.Equal(t, int64(0), m.rollbacksAccepted.Count())
		m.RollbackAccepted()
		assert.Equal(t, int64(1), m.rollbacksAccepted.Count())

		require.Equal(t, float64(0), m.snapshotCronLatency.Percentile(0.95))
		m.TimeUpdateSaturatedSnapshots(func() {
			time.Sleep(100 * time.Millisecond)
		})
		assert.True(t, m.snapshotCronLatency.Percentile(0.95) >= 100000) // FIXME wrong too low value in ns
	}
}
