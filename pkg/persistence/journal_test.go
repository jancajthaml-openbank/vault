package persistence

import (
	"io/ioutil"
	"math"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jancajthaml-openbank/vault/pkg/model"
	"github.com/jancajthaml-openbank/vault/pkg/utils"

	money "gopkg.in/inf.v0"
)

var journalTestParams = utils.RunParams{
	Setup: utils.SetupParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	},
	Journal: utils.JournalParams{},
	Metrics: utils.MetricsParams{},
}

func TestSnapshotUpdate(t *testing.T) {
	name := "account_2"
	currency := "XRP"
	isBalanceCheck := false

	snapshotInitial := CreateAccount(journalTestParams, name, currency, isBalanceCheck)
	require.NotNil(t, snapshotInitial)

	loadedInitial := LoadAccount(journalTestParams, name)
	require.NotNil(t, loadedInitial)

	t.Log("Initial matches loaded")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedInitial.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedInitial.Promised)
		assert.Equal(t, snapshotInitial.PromiseBuffer, loadedInitial.PromiseBuffer)
		assert.Equal(t, snapshotInitial.Version, loadedInitial.Version)
	}

	snapshotVersion1 := UpdateAccount(journalTestParams, name, snapshotInitial)
	require.NotNil(t, snapshotVersion1)

	loadedVersion1 := LoadAccount(journalTestParams, name)
	require.NotNil(t, loadedVersion1)

	t.Log("Updated matches loaded")
	{
		assert.Equal(t, snapshotVersion1.Balance, loadedVersion1.Balance)
		assert.Equal(t, snapshotVersion1.Promised, loadedVersion1.Promised)
		assert.Equal(t, snapshotVersion1.PromiseBuffer, loadedVersion1.PromiseBuffer)
		assert.Equal(t, snapshotVersion1.Version, loadedVersion1.Version)
	}

	t.Log("Updated is increment of version of initial by 1")
	{
		assert.Equal(t, loadedInitial.Version+1, loadedVersion1.Version)
		assert.Equal(t, snapshotInitial.Version+1, snapshotVersion1.Version)
	}
}

func TestRefuseSnapshotOverflow(t *testing.T) {
	name := "xxx"
	currency := "XxX"
	isBalanceCheck := true

	snapshotLast := &model.Account{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       int(math.MaxInt32),
		AccountName: name,
		Currency: currency,
		IsBalanceCheck: isBalanceCheck,
	}

	snapshotNext := UpdateAccount(journalTestParams, name, snapshotLast)

	assert.Equal(t, snapshotLast.Version, snapshotNext.Version)
}

func TestSnapshotPromiseBuffer(t *testing.T) {
	name := "yyy"
	currency := "yYy"
	isBalanceCheck := false

	expectedPromises := []string{"A", "B", "C", "D"}

	snapshot := &model.Account{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       0,
		AccountName: name,
		Currency: currency,
		IsBalanceCheck: isBalanceCheck,
	}

	snapshot.PromiseBuffer.AddAll(expectedPromises)

	PersistAccount(journalTestParams, name, snapshot)

	loaded := LoadAccount(journalTestParams, name)

	if loaded == nil {
		t.Errorf("Expected to load snapshot got nil instead")
		return
	}

	assert.Equal(t, snapshot.Balance, loaded.Balance)
	assert.Equal(t, snapshot.Promised, loaded.Promised)
	assert.Equal(t, snapshot.Version, loaded.Version)
	assert.Equal(t, snapshot.AccountName, loaded.AccountName)
	assert.Equal(t, snapshot.Currency, loaded.Currency)
	assert.Equal(t, snapshot.IsBalanceCheck, loaded.IsBalanceCheck)

	require.Equal(t, len(expectedPromises), loaded.PromiseBuffer.Size())
	for _, v := range expectedPromises {
		assert.True(t, snapshot.PromiseBuffer.Contains(v))
	}
}

func BenchmarkAccountLoad(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		b.Fatal(err.Error())
	}
	defer os.RemoveAll(tmpDir)

	params := utils.RunParams{
		Setup: utils.SetupParams{
			Tenant:      "tenant",
			RootStorage: tmpDir,
		},
		Journal: utils.JournalParams{},
		Metrics: utils.MetricsParams{},
	}

	account := CreateAccount(params, "bench", "BNC", false)
	require.NotNil(b, account)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadAccount(params, "bench")
	}
}
