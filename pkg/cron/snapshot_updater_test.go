package cron

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jancajthaml-openbank/vault/pkg/actor"
	"github.com/jancajthaml-openbank/vault/pkg/metrics"
	"github.com/jancajthaml-openbank/vault/pkg/model"
	"github.com/jancajthaml-openbank/vault/pkg/persistence"
	"github.com/jancajthaml-openbank/vault/pkg/utils"

	money "gopkg.in/inf.v0"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var snapshotTestParams = utils.RunParams{
	Setup: utils.SetupParams{
		Tenant:      "tenant",
		RootStorage: "/tmp/cron",
	},
	Journal: utils.JournalParams{
		JournalSaturation: 1,
	},
	Metrics: utils.MetricsParams{},
}

var testMetrics = metrics.NewMetrics()

func init() {
	removeContents := func(dir string) {
		d, err := os.Open(dir)
		if err != nil {
			return
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			return
		}
		for _, name := range names {
			err = os.RemoveAll(filepath.Join(dir, name))
			if err != nil {
				return
			}
		}
		return
	}

	removeContents(snapshotTestParams.Setup.RootStorage)
}

func TestUpdateSaturated(t *testing.T) {

	//require.NotNil(t, persistence.CreateAccount(snapshotTestParams, "account_1", "EUR", true))
	s := persistence.CreateAccount(snapshotTestParams, "account_1", "EUR", true)
	require.NotNil(t, s)
	require.True(t, persistence.PersistPromise(snapshotTestParams, "account_1", 0, new(money.Dec), "transaction_1"))
	s = persistence.UpdateAccount(snapshotTestParams, "account_1", s)
	require.True(t, persistence.PersistPromise(snapshotTestParams, "account_1", 1, new(money.Dec), "transaction_2"))
	require.True(t, persistence.PersistCommit(snapshotTestParams, "account_1", 1, new(money.Dec), "transaction_2"))
	require.NotNil(t, s)
	//require.NotNil(t, persistence.CreateMetadata(snapshotTestParams, "account_2", "EUR", true))
	require.NotNil(t, persistence.CreateAccount(snapshotTestParams, "account_2", "EUR", true))

	t.Log("return valid accounts")
	{
		assert.Equal(t, []string{"account_1", "account_2"}, getAccounts(snapshotTestParams))
	}

	t.Log("return valid version")
	{
		assert.Equal(t, 1, getVersion(snapshotTestParams, "account_1"))
		assert.Equal(t, 0, getVersion(snapshotTestParams, "account_2"))
		assert.Equal(t, -1, getVersion(snapshotTestParams, "account_3"))
	}

	t.Log("return valid events")
	{
		assert.Equal(t, 1, getEvents(snapshotTestParams, "account_1", 0))
		assert.Equal(t, 2, getEvents(snapshotTestParams, "account_1", 1))
		assert.Equal(t, -1, getEvents(snapshotTestParams, "account_2", 0))
		assert.Equal(t, -1, getEvents(snapshotTestParams, "account_3", 0))
	}

	t.Log("updates expected accounts")
	{
		var callbackCalled = 0

		updateSaturated(snapshotTestParams, testMetrics, func(p utils.RunParams, m *metrics.Metrics, msg interface{}, account string, sender actor.Coordinates) {
			callbackCalled++
			assert.Equal(t, "account_1", account)
			switch m := msg.(type) {

			case model.Update:
				assert.Equal(t, 1, m.Version)

			default:
				t.Error("invalid message received in callback")

			}

		})

		if callbackCalled != 1 {
			t.Error("callback was not called")
		}
	}
}
