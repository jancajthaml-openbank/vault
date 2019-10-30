package metrics

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Persist(), "cannot persist nil reference")
	}

	t.Log("error when marshalling fails")
	{
		entity := Metrics{}
		assert.EqualError(t, entity.Persist(), "cannot marshall nil references")
	}

	t.Log("error when cannot open tempfile for writing")
	{
		entity := Metrics{
			output:              "/sys/kernel/security",
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		assert.NotNil(t, entity.Persist())
	}

	t.Log("happy path")
	{
		tmpfile, err := ioutil.TempFile(os.TempDir(), "test_metrics_persist")

		require.Nil(t, err)
		defer os.Remove(tmpfile.Name())

		entity := Metrics{
			output:              tmpfile.Name(),
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile(tmpfile.Name())
		require.Nil(t, err)

		assert.Equal(t, expected, actual)
	}
}

func TestHydrate(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Hydrate(), "cannot hydrate nil reference")
	}

	t.Log("happy path")
	{
		tmpfile, err := ioutil.TempFile(os.TempDir(), "test_metrics_hydrate")

		require.Nil(t, err)
		defer os.Remove(tmpfile.Name())

		old := Metrics{
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		old.snapshotCronLatency.Update(time.Duration(1))
		old.updatedSnapshots.Mark(2)
		old.createdAccounts.Inc(3)
		old.promisesAccepted.Inc(4)
		old.commitsAccepted.Inc(5)
		old.rollbacksAccepted.Inc(6)

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile(tmpfile.Name(), data, 0444))

		entity := Metrics{
			output:              tmpfile.Name(),
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Hydrate())

		assert.Equal(t, float64(1), entity.snapshotCronLatency.Percentile(0.95))
		assert.Equal(t, int64(2), entity.updatedSnapshots.Count())
		assert.Equal(t, int64(3), entity.createdAccounts.Count())
		assert.Equal(t, int64(4), entity.promisesAccepted.Count())
		assert.Equal(t, int64(5), entity.commitsAccepted.Count())
		assert.Equal(t, int64(6), entity.rollbacksAccepted.Count())
	}
}
