package persistence

import (
	"io/ioutil"
	"math"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localfs "github.com/jancajthaml-openbank/local-fs"
	money "gopkg.in/inf.v0"

	"github.com/jancajthaml-openbank/vault-unit/model"
)

func TestSnapshot_Update(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := localfs.NewPlaintextStorage(tmpdir)

	name := "account_name"
	format := "account_format"
	currency := "XRP"
	isBalanceCheck := false

	snapshotInitial := CreateAccount(&storage, name, format, currency, isBalanceCheck)
	require.NotNil(t, snapshotInitial)

	loadedInitial := LoadAccount(&storage, name)
	require.NotNil(t, loadedInitial)

	t.Log("Initial matches loaded")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedInitial.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedInitial.Promised)
		assert.Equal(t, snapshotInitial.PromiseBuffer, loadedInitial.PromiseBuffer)
		assert.Equal(t, snapshotInitial.Version, loadedInitial.Version)
	}

	snapshotVersion1 := UpdateAccount(&storage, name, snapshotInitial)
	require.NotNil(t, snapshotVersion1)

	loadedVersion1 := LoadAccount(&storage, name)
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

func TestSnapshot_RefuseOverflow(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := localfs.NewPlaintextStorage(tmpdir)

	name := "xxx"
	format := "format"
	currency := "XxX"
	isBalanceCheck := true

	snapshotLast := &model.Account{
		Balance:        new(money.Dec),
		Promised:       new(money.Dec),
		PromiseBuffer:  model.NewTransactionSet(),
		Version:        int64(math.MaxInt32),
		Name:           name,
		Format:         format,
		Currency:       currency,
		IsBalanceCheck: isBalanceCheck,
	}

	snapshotNext := UpdateAccount(&storage, name, snapshotLast)

	assert.Equal(t, snapshotLast.Version, snapshotNext.Version)
}

func TestSnapshot_PromiseBuffer(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	storage := localfs.NewPlaintextStorage(tmpdir)

	name := "yyy"
	format := "format"
	currency := "yYy"
	isBalanceCheck := false

	expectedPromises := []string{"A", "B", "C", "D"}

	var snapshot = &model.Account{
		Balance:        new(money.Dec),
		Promised:       new(money.Dec),
		PromiseBuffer:  model.NewTransactionSet(),
		Version:        -1,
		Name:           name,
		Format:         format,
		Currency:       currency,
		IsBalanceCheck: isBalanceCheck,
	}
	snapshot.PromiseBuffer.Add(expectedPromises...)
	snapshot = UpdateAccount(&storage, name, snapshot)

	loaded := LoadAccount(&storage, name)

	if loaded == nil {
		t.Errorf("Expected to load snapshot got nil instead")
		return
	}

	assert.Equal(t, snapshot.Balance, loaded.Balance)
	assert.Equal(t, snapshot.Promised, loaded.Promised)
	assert.Equal(t, snapshot.Version, loaded.Version)
	assert.Equal(t, snapshot.Name, loaded.Name)
	assert.Equal(t, snapshot.Format, loaded.Format)
	assert.Equal(t, snapshot.Currency, loaded.Currency)
	assert.Equal(t, snapshot.IsBalanceCheck, loaded.IsBalanceCheck)

	require.Equal(t, len(expectedPromises), loaded.PromiseBuffer.Size())
	for _, v := range expectedPromises {
		assert.True(t, snapshot.PromiseBuffer.Contains(v))
	}
}

func BenchmarkAccountLoad(b *testing.B) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test_storage")
	require.Nil(b, err)
	defer os.RemoveAll(tmpdir)

	storage := localfs.NewPlaintextStorage(tmpdir)

	account := CreateAccount(&storage, "bench", "format", "BNC", false)
	require.NotNil(b, account)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadAccount(&storage, "bench")
	}
}
