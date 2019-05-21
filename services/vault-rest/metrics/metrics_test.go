package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilename(t *testing.T) {
	assert.Equal(t, "/a/b/c.e", getFilename("/a/b/c.e"))
}

func TestMetricsPersist(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, "", time.Hour)
	delay := 1e8
	delta := 1e8

	t.Log("TimeGetAccount properly times run of GetAccount function")
	{
		require.Equal(t, int64(0), entity.getAccountLatency.Count())
		entity.TimeGetAccount(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.getAccountLatency.Count())
		assert.InDelta(t, entity.getAccountLatency.Percentile(0.95), delay, delta)
	}

	t.Log("TimeCreateAccount properly times run of CreateAccount function")
	{
		require.Equal(t, int64(0), entity.createAccountLatency.Count())
		entity.TimeCreateAccount(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.createAccountLatency.Count())
		assert.InDelta(t, entity.createAccountLatency.Percentile(0.95), delay, delta)
	}
}
