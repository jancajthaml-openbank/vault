package metrics

import (
	"encoding/json"
	"testing"
	"time"

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
