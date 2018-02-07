package utils

import (
	"math"
	"os"
	"path/filepath"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jancajthaml-openbank/vault/model"

	money "gopkg.in/inf.v0"
)

func TestMetadataPersist(t *testing.T) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	name := "account_1"
	currency := "XRP"
	isBalanceCheck := false

	meta := CreateMetadata(params, name, currency, isBalanceCheck)
	require.NotNil(t, meta)

	loaded := LoadMetadata(params, name)
	require.NotNil(t, loaded)

	assert.Equal(t, meta.AccountName, loaded.AccountName)
	assert.Equal(t, name, loaded.AccountName)
	assert.Equal(t, meta.Currency, loaded.Currency)
	assert.Equal(t, currency, loaded.Currency)
	assert.Equal(t, meta.IsBalanceCheck, loaded.IsBalanceCheck)
	assert.Equal(t, isBalanceCheck, loaded.IsBalanceCheck)
}

func TestSnapshotPersist(t *testing.T) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	name := "account_1"

	snapshot := CreateSnapshot(params, name)
	require.NotNil(t, snapshot)

	loaded := LoadSnapshot(params, name)
	require.NotNil(t, loaded)

	assert.Equal(t, 0, loaded.Balance.Sign())
	assert.Equal(t, 0, loaded.Promised.Sign())
	assert.Equal(t, 0, loaded.Version)
}

func TestSnapshotUpdate(t *testing.T) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	name := "account_2"

	snapshotInitial := CreateSnapshot(params, name)
	require.NotNil(t, snapshotInitial)

	loadedInitial := LoadSnapshot(params, name)
	require.NotNil(t, loadedInitial)

	t.Log("Initial matches loaded")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedInitial.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedInitial.Promised)
		assert.Equal(t, snapshotInitial.PromiseBuffer, loadedInitial.PromiseBuffer)
		assert.Equal(t, snapshotInitial.Version, loadedInitial.Version)
	}

	snapshotVersion1 := UpdateSnapshot(params, name, snapshotInitial)
	require.NotNil(t, snapshotVersion1)

	loadedVersion1 := LoadSnapshot(params, name)
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
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	name := "xxx"

	snapshotLast := &model.Snapshot{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       int(math.MaxInt32),
	}

	snapshotNext := UpdateSnapshot(params, name, snapshotLast)

	assert.Equal(t, snapshotLast.Version, snapshotNext.Version)
}

func TestSnapshotPromiseBuffer(t *testing.T) {
	params := RunParams{
		Tenant:      "tenant",
		RootStorage: "/tmp",
	}
	name := "account_3"

	expectedPromises := []string{"A", "B", "C", "D"}

	snapshot := &model.Snapshot{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       0,
	}

	snapshot.PromiseBuffer.AddAll(expectedPromises)

	StoreSnapshot(params, name, snapshot)

	loaded := LoadSnapshot(params, name)

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

var benchmarkParams = RunParams{
	Tenant:      "tenant",
	RootStorage: "/benchmark",
}

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

	removeContents(benchmarkParams.RootStorage)

	meta := CreateMetadata(benchmarkParams, "bench", "cur", false)
	if meta == nil {
		panic("nil meta")
	}

	snapshot := CreateSnapshot(benchmarkParams, "bench")
	if snapshot == nil {
		panic("nil snapshot")
	}
}

func BenchmarkMetadataLoad(b *testing.B) {

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadMetadata(benchmarkParams, "bench")
	}
}

func BenchmarkSnapshotLoad(b *testing.B) {

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadSnapshot(benchmarkParams, "bench")
	}
}
