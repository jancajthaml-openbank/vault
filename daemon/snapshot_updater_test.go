package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jancajthaml-openbank/vault/actor"
	"github.com/jancajthaml-openbank/vault/config"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/persistence"

	money "gopkg.in/inf.v0"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func removeContents(dir string) {
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

type CallbackMessage struct {
	msg     interface{}
	account string
}

func TestSnapshotUpdater(t *testing.T) {
	cfg := config.Configuration{
		Tenant:            "tenant",
		RootStorage:       "/tmp/cron",
		JournalSaturation: 1,
	}

	defer removeContents(cfg.RootStorage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callbackCalled := 0
	callbackBacklog := make([]CallbackMessage, 0)

	callback := func(msg interface{}, account string, sender actor.Coordinates) {
		callbackBacklog = append(callbackBacklog, CallbackMessage{
			msg:     msg,
			account: account,
		})
		callbackCalled++
	}

	metrics := NewMetrics(ctx, cfg)
	su := NewSnapshotUpdater(ctx, cfg, &metrics, callback)

	s := persistence.CreateAccount(cfg.RootStorage, "account_1", "EUR", true)
	require.NotNil(t, s)
	require.True(t, persistence.PersistPromise(cfg.RootStorage, "account_1", 0, new(money.Dec), "transaction_1"))
	s = persistence.UpdateAccount(cfg.RootStorage, "account_1", s)
	require.True(t, persistence.PersistPromise(cfg.RootStorage, "account_1", 1, new(money.Dec), "transaction_2"))
	require.True(t, persistence.PersistCommit(cfg.RootStorage, "account_1", 1, new(money.Dec), "transaction_2"))
	require.NotNil(t, s)

	require.NotNil(t, persistence.CreateAccount(cfg.RootStorage, "account_2", "EUR", true))

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
