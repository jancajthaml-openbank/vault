package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	money "gopkg.in/inf.v0"
)

/*
func TestAccount_Copy(t *testing.T) {
	t.Log("can copy nil")
	{
		var entity *Account
		assert.Equal(t, entity.Copy(), entity)
	}
}
*/

func TestAccount_Serialize(t *testing.T) {
	t.Log("serialized is deserializable")
	{
		entity := new(Account)

		entity.Name = "accountName"
		entity.Format = "FOR"
		entity.Currency = "CUR"
		entity.IsBalanceCheck = false
		entity.SnapshotVersion = 3

		var ok bool

		entity.Balance, ok = new(money.Dec).SetString("1.0")
		require.True(t, ok)

		entity.Promised, ok = new(money.Dec).SetString("2.0")
		require.True(t, ok)

		entity.Promises = NewPromises()
		entity.Promises.Add("A", "B", "C", "D", "E", "F", "G", "H")

		data := entity.Serialize()
		require.NotNil(t, data)

		assert.Equal(t, "CUR FOR_F\n1.0\n2.0\nA\nB\nC\nD\nE\nF\nG\nH", string(data))

		hydrated := new(Account)
		hydrated.Name = entity.Name
		hydrated.SnapshotVersion = entity.SnapshotVersion

		hydrated.Deserialize(data)

		assert.Equal(t, entity, hydrated)
	}

	t.Log("serialized is initialState")
	{
		entity := new(Account)
		entity.IsBalanceCheck = true
		entity.Format = "FOR"
		entity.Currency = "CUR"

		data := entity.Serialize()
		require.NotNil(t, data)

		assert.Equal(t, data, []byte("CUR FOR_T\n0.0\n0.0"))
	}

	t.Log("serialized isBalanceCheck")
	{
		yes := new(Account)
		yes.IsBalanceCheck = true
		assert.Equal(t, yes.Serialize(), []byte("??? _T\n0.0\n0.0"))

		no := new(Account)
		no.IsBalanceCheck = false
		assert.Equal(t, no.Serialize(), []byte("??? _F\n0.0\n0.0"))
	}
}

func BenchmarkAccount_Serialize(b *testing.B) {
	entity := new(Account)

	entity.Name = "accountName"
	entity.Format = "accountFormat"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false
	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.Promises = NewPromises()
	entity.Promises.Add("A", "B", "C", "D", "E", "F", "G", "H")
	entity.SnapshotVersion = 0

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Serialize()
	}
}

func BenchmarkAccount_Deserialize(b *testing.B) {
	entity := new(Account)

	entity.Name = "accountName"
	entity.Format = "accountFormat"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false

	entity.Balance = new(money.Dec)
	entity.Promised = new(money.Dec)
	entity.Promises = NewPromises()
	entity.Promises.Add("A", "B", "C", "D", "E", "F", "G", "H")
	entity.SnapshotVersion = 0

	data := entity.Serialize()
	hydrated := new(Account)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		hydrated.Deserialize(data)
	}
}
