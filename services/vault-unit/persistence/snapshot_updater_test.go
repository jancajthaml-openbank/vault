package persistence

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jancajthaml-openbank/vault-unit/metrics"
	"github.com/jancajthaml-openbank/vault-unit/model"

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
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := localfs.NewStorage(tmpdir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callbackCalled := 0
	callbackBacklog := make([]CallbackMessage, 0)

	callback := func(msg interface{}, account system.Coordinates, sender system.Coordinates) {
		callbackBacklog = append(callbackBacklog, CallbackMessage{
			msg:     msg,
			account: account.Name,
		})
		callbackCalled++
	}

	metrics := metrics.NewMetrics(ctx, "tenant", "", time.Hour)
	su := NewSnapshotUpdater(ctx, 1, time.Hour, &metrics, &storage, callback)

	s := CreateAccount(&storage, "account_1", "EUR", true)
	require.NotNil(t, s)
	require.True(t, PersistPromise(&storage, "account_1", 0, new(money.Dec), "transaction_1"))
	s = UpdateAccount(&storage, "account_1", s)
	require.True(t, PersistPromise(&storage, "account_1", 1, new(money.Dec), "transaction_2"))
	require.True(t, PersistCommit(&storage, "account_1", 1, new(money.Dec), "transaction_2"))
	require.NotNil(t, s)

	require.NotNil(t, CreateAccount(&storage, "account_2", "EUR", true))

	t.Log("return valid accounts")
	{
		assert.Equal(t, []string{"account_1", "account_2"}, su.getAccounts())
	}

	t.Log("return valid version")
	{
		assert.Equal(t, 1, su.getVersion("account_1"))
		assert.Equal(t, 0, su.getVersion("account_2"))
		assert.Equal(t, -1, su.getVersion("account_3"))
	}

	t.Log("return valid events")
	{
		assert.Equal(t, 1, su.getEvents("account_1", 0))
		assert.Equal(t, 2, su.getEvents("account_1", 1))
		assert.Equal(t, -1, su.getEvents("account_2", 0))
		assert.Equal(t, -1, su.getEvents("account_3", 0))
	}

	t.Log("updates expected accounts")
	{
		su.updateSaturated()
		assert.Equal(t, 1, callbackCalled)
		assert.Equal(t, 1, len(callbackBacklog))

		args := callbackBacklog[0]
		assert.Equal(t, "account_1", args.account)
		switch m := args.msg.(type) {

		case model.Update:
			assert.Equal(t, 1, m.Version)

		default:
			t.Error("invalid message received in callback")

		}
	}
}
