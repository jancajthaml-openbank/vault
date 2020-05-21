package actor

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jancajthaml-openbank/vault-unit/metrics"
	"github.com/jancajthaml-openbank/vault-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
	money "gopkg.in/inf.v0"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CallbackMessage struct {
	msg     interface{}
	account string
}

func TestSnapshotUpdater(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "snapshot_test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := localfs.NewPlaintextStorage(tmpdir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callbackCalled := 0
	callbackBacklog := make([]CallbackMessage, 0)

	callback := func(msg string, account system.Coordinates, sender system.Coordinates) {
		callbackBacklog = append(callbackBacklog, CallbackMessage{
			msg:     msg,
			account: account.Name,
		})
		callbackCalled++
	}

	metrics := metrics.NewMetrics(ctx, "/tmp", "1", time.Hour)
	su := NewSnapshotUpdater(ctx, 1, time.Hour, &metrics, &storage, callback)

	s := persistence.CreateAccount(&storage, "s_account_1", "format", "EUR", true)
	require.NotNil(t, s)
	require.Nil(t, persistence.PersistPromise(&storage, "s_account_1", 0, new(money.Dec), "transaction_1"))
	s = persistence.UpdateAccount(&storage, "s_account_1", s)
	require.Nil(t, persistence.PersistPromise(&storage, "s_account_1", 1, new(money.Dec), "transaction_2"))
	require.Nil(t, persistence.PersistCommit(&storage, "s_account_1", 1, new(money.Dec), "transaction_2"))
	require.NotNil(t, s)

	require.NotNil(t, persistence.CreateAccount(&storage, "s_account_2", "format", "EUR", true))

	t.Log("return valid accounts")
	{
		assert.Equal(t, []string{"s_account_1", "s_account_2"}, su.getAccounts())
	}

	t.Log("return valid version")
	{
		assert.Equal(t, int64(1), su.getVersion("s_account_1"))
		assert.Equal(t, int64(0), su.getVersion("s_account_2"))
		assert.Equal(t, int64(-1), su.getVersion("s_account_3"))
	}

	t.Log("return valid events")
	{
		assert.Equal(t, 1, su.getEvents("s_account_1", 0))
		assert.Equal(t, 2, su.getEvents("s_account_1", 1))
		assert.Equal(t, -1, su.getEvents("s_account_2", 0))
		assert.Equal(t, -1, su.getEvents("s_account_3", 0))
	}

	t.Log("updates expected accounts")
	{
		su.updateSaturated()
		assert.Equal(t, 1, callbackCalled)
		assert.Equal(t, 1, len(callbackBacklog))

		args := callbackBacklog[0]
		assert.Equal(t, "s_account_1", args.account)
		assert.Equal(t, "US 1", args.msg)
	}
}
