package persistence

import (
	"io/ioutil"
	"math"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localfs "github.com/jancajthaml-openbank/local-fs"

	"github.com/jancajthaml-openbank/vault-unit/model"
)

func TestSnapshot_Update(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage, _ := localfs.NewPlaintextStorage(tmpdir)

	name := "account_name"
	format := "account_format"
	currency := "XRP"
	isBalanceCheck := false

	snapshotInitial, err := CreateAccount(storage, name, format, currency, isBalanceCheck)
	require.Nil(t, err)
	require.NotNil(t, snapshotInitial)

	loadedInitial, err := LoadAccount(storage, name)
	require.Nil(t, err)
	require.NotNil(t, loadedInitial)

	t.Log("Initial matches loaded")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedInitial.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedInitial.Promised)
		assert.Equal(t, snapshotInitial.Promises, loadedInitial.Promises)
		assert.Equal(t, snapshotInitial.SnapshotVersion, loadedInitial.SnapshotVersion)
	}

	err = UpdateAccount(storage, name, snapshotInitial)
	assert.Nil(t, err)

	loadedVersion1, err := LoadAccount(storage, name)
	require.Nil(t, err)
	require.NotNil(t, loadedVersion1)

	t.Log("Updated is increment of version of initial by 1")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedVersion1.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedVersion1.Promised)
		assert.Equal(t, snapshotInitial.Promises, loadedVersion1.Promises)
		assert.Equal(t, snapshotInitial.SnapshotVersion+1, loadedVersion1.SnapshotVersion)
		assert.Equal(t, int64(0), loadedVersion1.EventCounter)
	}
}

func TestSnapshot_RefuseOverflow(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage, _ := localfs.NewPlaintextStorage(tmpdir)

	name := "xxx"
	format := "format"
	currency := "XxX"
	isBalanceCheck := true

	snapshotLast := &model.Account{
		Balance:         new(model.Dec),
		Promised:        new(model.Dec),
		Promises:        model.NewPromises(),
		SnapshotVersion: int64(math.MaxInt32),
		EventCounter:    0,
		Name:            name,
		Format:          format,
		Currency:        currency,
		IsBalanceCheck:  isBalanceCheck,
	}

	err = UpdateAccount(storage, name, snapshotLast)
	require.NotNil(t, err)
}

func TestSnapshot_Promises(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage, _ := localfs.NewPlaintextStorage(tmpdir)

	name := "yyy"
	format := "format"
	currency := "yYy"
	isBalanceCheck := false

	expectedPromises := []string{"A", "B", "C", "D"}

	var snapshot = &model.Account{
		Balance:         new(model.Dec),
		Promised:        new(model.Dec),
		Promises:        model.NewPromises(),
		SnapshotVersion: 0,
		Name:            name,
		Format:          format,
		Currency:        currency,
		IsBalanceCheck:  isBalanceCheck,
	}
	snapshot.Promises.Add(expectedPromises...)

	err = UpdateAccount(storage, name, snapshot)
	require.Nil(t, err)

	loaded, err := LoadAccount(storage, name)
	require.Nil(t, err)

	if loaded == nil {
		t.Errorf("Expected to load snapshot got nil instead")
		return
	}

	assert.Equal(t, snapshot.Balance, loaded.Balance)
	assert.Equal(t, snapshot.Promised, loaded.Promised)
	assert.Equal(t, snapshot.SnapshotVersion+1, loaded.SnapshotVersion)
	assert.Equal(t, snapshot.Name, loaded.Name)
	assert.Equal(t, snapshot.Format, loaded.Format)
	assert.Equal(t, snapshot.Currency, loaded.Currency)
	assert.Equal(t, snapshot.IsBalanceCheck, loaded.IsBalanceCheck)

	require.Equal(t, len(expectedPromises), loaded.Promises.Size())
	for _, v := range expectedPromises {
		assert.True(t, snapshot.Promises.Contains(v))
	}
}

func BenchmarkAccountLoad(b *testing.B) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(b, err)
	defer os.RemoveAll(tmpdir)

	storage, _ := localfs.NewPlaintextStorage(tmpdir)

	account, err := CreateAccount(storage, "bench", "format", "BNC", false)
	require.Nil(b, err)
	require.NotNil(b, account)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadAccount(storage, "bench")
	}
}
