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

func TestMetadataPersist(t *testing.T) {
	name := "account_1"
	currency := "XRP"
	isBalanceCheck := false

	meta := CreateMetadata(journalTestParams, name, currency, isBalanceCheck)
	require.NotNil(t, meta)

	loaded := LoadMetadata(journalTestParams, name)
	require.NotNil(t, loaded)

	assert.Equal(t, meta.AccountName, loaded.AccountName)
	assert.Equal(t, name, loaded.AccountName)
	assert.Equal(t, meta.Currency, loaded.Currency)
	assert.Equal(t, currency, loaded.Currency)
	assert.Equal(t, meta.IsBalanceCheck, loaded.IsBalanceCheck)
	assert.Equal(t, isBalanceCheck, loaded.IsBalanceCheck)
}

func TestSnapshotPersist(t *testing.T) {
	name := "account_1"

	snapshot := CreateSnapshot(journalTestParams, name)
	require.NotNil(t, snapshot)

	loaded := LoadSnapshot(journalTestParams, name)
	require.NotNil(t, loaded)

	assert.Equal(t, 0, loaded.Balance.Sign())
	assert.Equal(t, 0, loaded.Promised.Sign())
	assert.Equal(t, 0, loaded.Version)
}

func TestSnapshotUpdate(t *testing.T) {
	name := "account_2"

	snapshotInitial := CreateSnapshot(journalTestParams, name)
	require.NotNil(t, snapshotInitial)

	loadedInitial := LoadSnapshot(journalTestParams, name)
	require.NotNil(t, loadedInitial)

	t.Log("Initial matches loaded")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedInitial.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedInitial.Promised)
		assert.Equal(t, snapshotInitial.PromiseBuffer, loadedInitial.PromiseBuffer)
		assert.Equal(t, snapshotInitial.Version, loadedInitial.Version)
	}

	snapshotVersion1 := UpdateSnapshot(journalTestParams, name, snapshotInitial)
	require.NotNil(t, snapshotVersion1)

	loadedVersion1 := LoadSnapshot(journalTestParams, name)
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

	snapshotLast := &model.Snapshot{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       int(math.MaxInt32),
	}

	snapshotNext := UpdateSnapshot(journalTestParams, name, snapshotLast)

	assert.Equal(t, snapshotLast.Version, snapshotNext.Version)
}

func TestSnapshotPromiseBuffer(t *testing.T) {
	name := "account_3"

	expectedPromises := []string{"A", "B", "C", "D"}

	snapshot := &model.Snapshot{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       0,
	}

	snapshot.PromiseBuffer.AddAll(expectedPromises)

	StoreSnapshot(journalTestParams, name, snapshot)

	loaded := LoadSnapshot(journalTestParams, name)

	if loaded == nil {
		t.Errorf("Expected to load snapshot got nil instead")
		return
	}

	assert.Equal(t, snapshot.Balance, loaded.Balance)
	assert.Equal(t, snapshot.Promised, loaded.Promised)
	assert.Equal(t, snapshot.Version, loaded.Version)

	require.Equal(t, len(expectedPromises), loaded.PromiseBuffer.Size())
	for _, v := range expectedPromises {
		assert.True(t, snapshot.PromiseBuffer.Contains(v))
	}
}

func BenchmarkMetadataLoad(b *testing.B) {
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

	meta := CreateMetadata(params, "bench", "cur", false)
	require.NotNil(b, meta)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadMetadata(params, "bench")
	}
}

func BenchmarkSnapshotLoad(b *testing.B) {
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

	snapshot := CreateSnapshot(params, "bench")
	require.NotNil(b, snapshot)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadSnapshot(params, "bench")
	}
}
