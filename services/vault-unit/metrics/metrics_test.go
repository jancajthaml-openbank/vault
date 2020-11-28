package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	entity := NewMetrics("/tmp", "1")
	delay := 1e8
	delta := 1e8

	t.Log("TimeUpdateSaturatedSnapshots properly times run of UpdateSaturatedSnapshots function")
	{
		require.Equal(t, int64(0), entity.snapshotCronLatency.Count())
		entity.TimeUpdateSaturatedSnapshots(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.snapshotCronLatency.Count())
		assert.InDelta(t, entity.snapshotCronLatency.Percentile(0.95), delay, delta)
	}

	t.Log("SnapshotsUpdated properly updates number of updated snapshots")
	{
		require.Equal(t, int64(0), entity.updatedSnapshots.Count())
		entity.SnapshotsUpdated(2)
		assert.Equal(t, int64(2), entity.updatedSnapshots.Count())
	}

	t.Log("AccountCreated properly increments number of created accounts")
	{
		require.Equal(t, int64(0), entity.createdAccounts.Count())
		entity.AccountCreated()
		assert.Equal(t, int64(1), entity.createdAccounts.Count())
	}

	t.Log("PromiseAccepted properly increments number of accepted promises")
	{
		require.Equal(t, int64(0), entity.promisesAccepted.Count())
		entity.PromiseAccepted()
		assert.Equal(t, int64(1), entity.promisesAccepted.Count())
	}

	t.Log("CommitAccepted properly increments number of accepted commits")
	{
		require.Equal(t, int64(0), entity.commitsAccepted.Count())
		entity.CommitAccepted()
		assert.Equal(t, int64(1), entity.commitsAccepted.Count())
	}

	t.Log("RollbackAccepted properly increments number of accepted rollbacks")
	{
		require.Equal(t, int64(0), entity.rollbacksAccepted.Count())
		entity.RollbackAccepted()
		assert.Equal(t, int64(1), entity.rollbacksAccepted.Count())
	}
}
