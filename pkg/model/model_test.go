package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	money "gopkg.in/inf.v0"
)

func TestAccountSerializeForStorage(t *testing.T) {
	t.Log("serialized is deserializable")
	{
		entity := new(Account)

		entity.AccountName = "accountName"
		entity.Currency = "CUR"
		entity.IsBalanceCheck = false

		data := entity.SerializeForStorage()
		require.NotNil(t, data)

		hydrated := new(Account)
		hydrated.DeserializeFromStorage(data)

		assert.Equal(t, entity, hydrated)
	}

	t.Log("currency is normalized")
	{
		entity := new(Account)

		entity.Currency = "xXx"

		data := entity.SerializeForStorage()
		require.NotNil(t, data)

		hydrated := new(Account)
		hydrated.DeserializeFromStorage(data)

		assert.Equal(t, "XXX", hydrated.Currency)
	}
}

func TestStorageSerializeForStorage(t *testing.T) {
	t.Log("serialized is deserializable")
	{
		entity := new(Snapshot)

		var ok bool

		entity.Balance, ok = new(money.Dec).SetString("1.0")
		require.True(t, ok)

		entity.Promised, ok = new(money.Dec).SetString("2.0")
		require.True(t, ok)

		entity.PromiseBuffer = NewTransactionSet()
		entity.PromiseBuffer.AddAll([]string{"A", "B", "C", "D", "E", "F", "G", "H"})

		entity.Version = 3

		data := entity.SerializeForStorage()
		require.NotNil(t, data)

		hydrated := new(Snapshot)
		hydrated.DeserializeFromStorage(data)

		assert.Equal(t, entity, hydrated)
	}
}

func BenchmarkAccountSerializeForStorage(b *testing.B) {
	entity := new(Account)

	entity.AccountName = "accountName"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.SerializeForStorage()
	}
}

func BenchmarkAccountDeserializeFromStorage(b *testing.B) {
	entity := new(Account)

	entity.AccountName = "accountName"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false

	data := entity.SerializeForStorage()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		hydrated := new(Account)
		hydrated.DeserializeFromStorage(data)
	}
}

func BenchmarkSnapshotSerializeForStorage(b *testing.B) {
	entity := new(Snapshot)

	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.PromiseBuffer = NewTransactionSet()
	entity.PromiseBuffer.AddAll([]string{"A", "B", "C", "D", "E", "F", "G", "H"})
	entity.Version = 0

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.SerializeForStorage()
	}
}

func BenchmarkSnapshotDeserializeFromStorage(b *testing.B) {
	entity := new(Snapshot)

	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.PromiseBuffer = NewTransactionSet()
	entity.PromiseBuffer.AddAll([]string{"A", "B", "C", "D", "E", "F", "G", "H"})
	entity.Version = 0

	data := entity.SerializeForStorage()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		hydrated := new(Snapshot)
		hydrated.DeserializeFromStorage(data)
	}
}
