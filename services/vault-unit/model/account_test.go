package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccount_Deserialize(t *testing.T) {

	t.Log("does not panic on nil")
	{
		var entity *Account
		entity.Deserialize(nil)
	}

	t.Log("cut after meta")
	{
		data := []byte("CUR FOR_T")

		entity := new(Account)
		entity.Deserialize(data)

		assert.Equal(t, "FOR", entity.Format)
		assert.Equal(t, "CUR", entity.Currency)
		assert.Equal(t, true, entity.IsBalanceCheck)
		assert.Nil(t, entity.Balance)
		assert.Nil(t, entity.Promised)
		assert.Equal(t, int64(0), entity.SnapshotVersion)
		assert.Equal(t, int64(0), entity.EventCounter)
	}

	t.Log("cut after balance")
	{
		data := []byte("CUR FOR_T\n1.0")

		entity := new(Account)
		entity.Deserialize(data)

		assert.Equal(t, "FOR", entity.Format)
		assert.Equal(t, "CUR", entity.Currency)
		assert.Equal(t, true, entity.IsBalanceCheck)
		assert.NotNil(t, entity.Balance)
		assert.Equal(t, "1.0", entity.Balance.String())
		assert.Nil(t, entity.Promised)
		assert.Equal(t, int64(0), entity.SnapshotVersion)
		assert.Equal(t, int64(0), entity.EventCounter)
	}

	t.Log("cut after promised")
	{
		data := []byte("CUR FOR_T\n1.0\n2.0")

		entity := new(Account)
		entity.Deserialize(data)

		assert.Equal(t, "FOR", entity.Format)
		assert.Equal(t, "CUR", entity.Currency)
		assert.Equal(t, true, entity.IsBalanceCheck)
		assert.NotNil(t, entity.Balance)
		assert.Equal(t, "1.0", entity.Balance.String())
		assert.NotNil(t, entity.Promised)
		assert.Equal(t, "2.0", entity.Promised.String())
		assert.Equal(t, int64(0), entity.SnapshotVersion)
		assert.Equal(t, int64(0), entity.EventCounter)
	}

	t.Log("full")
	{
		data := []byte("CUR FOR_T\n1.0\n2.0\nA\nB")

		entity := new(Account)
		entity.Deserialize(data)

		assert.Equal(t, "FOR", entity.Format)
		assert.Equal(t, "CUR", entity.Currency)
		assert.Equal(t, true, entity.IsBalanceCheck)
		assert.NotNil(t, entity.Balance)
		assert.Equal(t, "1.0", entity.Balance.String())
		assert.NotNil(t, entity.Promised)
		assert.Equal(t, "2.0", entity.Promised.String())
		assert.Equal(t, "[A,B]", entity.Promises.String())
		assert.Equal(t, int64(0), entity.SnapshotVersion)
		assert.Equal(t, int64(0), entity.EventCounter)
	}
}

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

		entity.Balance, ok = new(Dec).SetString("1.0")
		require.True(t, ok)

		entity.Promised, ok = new(Dec).SetString("2.0")
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

	t.Log("serialized currency")
	{
		a := new(Account)
		a.Currency = ""
		assert.Equal(t, a.Serialize(), []byte("??? _F\n0.0\n0.0"))

		b := new(Account)
		b.Currency = "A"
		assert.Equal(t, b.Serialize(), []byte("A?? _F\n0.0\n0.0"))

		c := new(Account)
		c.Currency = "AB"
		assert.Equal(t, c.Serialize(), []byte("AB? _F\n0.0\n0.0"))

		d := new(Account)
		d.Currency = "ABC"
		assert.Equal(t, d.Serialize(), []byte("ABC _F\n0.0\n0.0"))

		e := new(Account)
		e.Currency = "ABCD"
		assert.Equal(t, e.Serialize(), []byte("ABC _F\n0.0\n0.0"))
	}
}

func BenchmarkAccount_Serialize(b *testing.B) {
	entity := new(Account)

	entity.Name = "accountName"
	entity.Format = "accountFormat"
	entity.Currency = "CUR"
	entity.IsBalanceCheck = false
	entity.Balance = new(Dec)
	entity.Promised = new(Dec)
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

	entity.Balance = new(Dec)
	entity.Promised = new(Dec)
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
