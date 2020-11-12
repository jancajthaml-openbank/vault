package metrics

import (
	"encoding/json"
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
		assert.EqualError(t, err, "cannot marshal nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshal nil references")
	}

	t.Log("happy path")
	{
		entity := Metrics{
			getAccountLatency:    metrics.NewTimer(),
			createAccountLatency: metrics.NewTimer(),
		}

		entity.getAccountLatency.Update(time.Duration(1))
		entity.createAccountLatency.Update(time.Duration(2))

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		aux := &struct {
			GetAccountLatency    float64 `json:"getAccountLatency"`
			CreateAccountLatency float64 `json:"createAccountLatency"`
		}{}

		require.Nil(t, json.Unmarshal(actual, &aux))

		assert.Equal(t, float64(1), aux.GetAccountLatency)
		assert.Equal(t, float64(2), aux.CreateAccountLatency)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshal to nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshal to nil references")
	}

	t.Log("error on malformed data")
	{
		entity := Metrics{
			getAccountLatency:    metrics.NewTimer(),
			createAccountLatency: metrics.NewTimer(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			getAccountLatency:    metrics.NewTimer(),
			createAccountLatency: metrics.NewTimer(),
		}

		data := []byte("{\"getAccountLatency\":1,\"createAccountLatency\":2}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, float64(1), entity.getAccountLatency.Percentile(0.95))
		assert.Equal(t, float64(2), entity.createAccountLatency.Percentile(0.95))
	}
}

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.NotNil(t, entity.Persist())
	}

	t.Log("error when marshaling fails")
	{
		entity := Metrics{}
		assert.NotNil(t, entity.Persist())
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		storage, _ := localfs.NewPlaintextStorage("/tmp")
		entity := Metrics{
			storage:              storage,
			getAccountLatency:    metrics.NewTimer(),
			createAccountLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile("/tmp/metrics.json")
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
		defer os.Remove("/tmp/metrics.json")

		old := Metrics{
			getAccountLatency:    metrics.NewTimer(),
			createAccountLatency: metrics.NewTimer(),
		}

		old.getAccountLatency.Update(time.Duration(1))
		old.createAccountLatency.Update(time.Duration(2))

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile("/tmp/metrics.json", data, 0444))

		storage, _ := localfs.NewPlaintextStorage("/tmp")
		entity := Metrics{
			storage:              storage,
			getAccountLatency:    metrics.NewTimer(),
			createAccountLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Hydrate())

		assert.Equal(t, float64(1), entity.getAccountLatency.Percentile(0.95))
		assert.Equal(t, float64(2), entity.createAccountLatency.Percentile(0.95))
	}
}
