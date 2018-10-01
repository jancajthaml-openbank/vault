package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	money "gopkg.in/inf.v0"
)

func TestAccountHydrate(t *testing.T) {
	t.Log("serialized is deserializable")
	{
		entity := new(Account)

		entity.AccountName = "accountName"
		entity.Currency = "CUR"
		entity.IsBalanceCheck = false

		var ok bool

		entity.Balance, ok = new(money.Dec).SetString("1.0")
		require.True(t, ok)

		entity.Promised, ok = new(money.Dec).SetString("2.0")
		require.True(t, ok)

		entity.PromiseBuffer = NewTransactionSet()
		entity.PromiseBuffer.AddAll([]string{"A", "B", "C", "D", "E", "F", "G", "H"})

		entity.Version = 3

		data := entity.Persist()
		require.NotNil(t, data)

		hydrated := new(Account)
		hydrated.Hydrate(data)

		assert.Equal(t, entity, hydrated)
	}
}

func BenchmarkAccountPersist(b *testing.B) {
	entity := new(Account)

	entity.AccountName = "accountName"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false
	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.PromiseBuffer = NewTransactionSet()
	entity.PromiseBuffer.AddAll([]string{"A", "B", "C", "D", "E", "F", "G", "H"})
	entity.Version = 0

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Persist()
	}
}

func BenchmarkAccountHydrate(b *testing.B) {
	entity := new(Account)

	entity.AccountName = "accountName"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false

	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.PromiseBuffer = NewTransactionSet()
	entity.PromiseBuffer.AddAll([]string{"A", "B", "C", "D", "E", "F", "G", "H"})
	entity.Version = 0

	data := entity.Persist()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		hydrated := new(Account)
		hydrated.Hydrate(data)
	}
}
