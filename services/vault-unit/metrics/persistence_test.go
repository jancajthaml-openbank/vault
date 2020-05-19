package metrics

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil references")
	}

	t.Log("happy path")
	{
		entity := Metrics{
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		entity.snapshotCronLatency.Update(time.Duration(1))
		entity.updatedSnapshots.Mark(2)
		entity.createdAccounts.Inc(3)
		entity.promisesAccepted.Inc(4)
		entity.commitsAccepted.Inc(5)
		entity.rollbacksAccepted.Inc(6)

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		data := []byte("{\"snapshotCronLatency\":1,\"updatedSnapshots\":2,\"createdAccounts\":3,\"promisesAccepted\":4,\"commitsAccepted\":5,\"rollbacksAccepted\":6}")

		assert.Equal(t, data, actual)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil references")
	}

	t.Log("error on malformed data")
	{
		entity := Metrics{
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			promisesAccepted:    metrics.NewCounter(),
			commitsAccepted:     metrics.NewCounter(),
			rollbacksAccepted:   metrics.NewCounter(),
			createdAccounts:     metrics.NewCounter(),
			updatedSnapshots:    metrics.NewMeter(),
			snapshotCronLatency: metrics.NewTimer(),
		}

		data := []byte("{\"snapshotCronLatency\":1,\"updatedSnapshots\":2,\"createdAccounts\":3,\"promisesAccepted\":4,\"commitsAccepted\":5,\"rollbacksAccepted\":6}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, float64(1), entity.snapshotCronLatency.Percentile(0.95))
		assert.Equal(t, int64(2), entity.updatedSnapshots.Count())
		assert.Equal(t, int64(3), entity.createdAccounts.Count())
		assert.Equal(t, int64(4), entity.promisesAccepted.Count())
		assert.Equal(t, int64(5), entity.commitsAccepted.Count())
		assert.Equal(t, int64(6), entity.rollbacksAccepted.Count())
	}
}

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

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		entity := Metrics{
			storage:             localfs.NewPlaintextStorage("/tmp"),
			tenant:              "1",
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

		actual, err := ioutil.ReadFile("/tmp/metrics.1.json")
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

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Hydrate(), "cannot hydrate nil reference")
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.1.json")

		old := Metrics{
			storage:             localfs.NewPlaintextStorage("/tmp"),
			tenant:              "1",
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

		require.Nil(t, ioutil.WriteFile("/tmp/metrics.1.json", data, 0444))

		entity := Metrics{
			storage:             localfs.NewPlaintextStorage("/tmp"),
			tenant:              "1",
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
