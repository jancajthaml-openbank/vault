package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	money "gopkg.in/inf.v0"
)

func TestAccount_Serialise(t *testing.T) {
	t.Log("serialized is deserializable")
	{
		entity := new(Account)

		entity.AccountName = "accountName"
		entity.Currency = "CUR"
		entity.IsBalanceCheck = false
		entity.Version = 3

		var ok bool

		entity.Balance, ok = new(money.Dec).SetString("1.0")
		require.True(t, ok)

		entity.Promised, ok = new(money.Dec).SetString("2.0")
		require.True(t, ok)

		entity.PromiseBuffer = NewTransactionSet()
		entity.PromiseBuffer.Add("A", "B", "C", "D", "E", "F", "G", "H")

		data := entity.Serialise()
		require.NotNil(t, data)

		assert.Equal(t, "FCUR\n1.0\n2.0\nA\nB\nC\nD\nE\nF\nG\nH", string(data))

		hydrated := new(Account)
		hydrated.AccountName = entity.AccountName
		hydrated.Version = entity.Version

		hydrated.Deserialise(data)

		assert.Equal(t, entity, hydrated)
	}

	t.Log("serialized is initialState")
	{
		entity := new(Account)
		entity.IsBalanceCheck = true
		entity.Currency = "CUR"

		data := entity.Serialise()
		require.NotNil(t, data)

		assert.Equal(t, data, []byte("TCUR\n0.0\n0.0"))
	}

	t.Log("serialized isBalanceCheck")
	{
		yes := new(Account)
		yes.IsBalanceCheck = true
		assert.Equal(t, yes.Serialise(), []byte("T???\n0.0\n0.0"))

		no := new(Account)
		no.IsBalanceCheck = false
		assert.Equal(t, no.Serialise(), []byte("F???\n0.0\n0.0"))
	}
}

func BenchmarkAccount_Serialise(b *testing.B) {
	entity := new(Account)

	entity.AccountName = "accountName"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false
	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.PromiseBuffer = NewTransactionSet()
	entity.PromiseBuffer.Add("A", "B", "C", "D", "E", "F", "G", "H")
	entity.Version = 0

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Serialise()
	}
}

func BenchmarkAccount_Deserialise(b *testing.B) {
	entity := new(Account)

	entity.AccountName = "accountName"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false

	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.PromiseBuffer = NewTransactionSet()
	entity.PromiseBuffer.Add("A", "B", "C", "D", "E", "F", "G", "H")
	entity.Version = 0

	data := entity.Serialise()
	hydrated := new(Account)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		hydrated.Deserialise(data)
	}
}
