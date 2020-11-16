package model

import (
  "testing"

  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
)

func TestAccount_Unmarshall(t *testing.T) {
  t.Log("does not panic on nil")
  {
    var entity *Account
    entity.UnmarshalJSON(nil)
  }

  t.Log("does not panic on invalid data")
  {
    data := []byte(`
      {
        "name": "NAME",
    `)
    require.NotNil(t, new(Account).UnmarshalJSON(data))
  }

  t.Log("full schema check")
  {
    data := []byte(`
      {
        "name": "NAME",
        "format": "FORMAT",
        "currency": "CUR",
        "isBalanceCheck": true
      }
    `)
    entity := new(Account)
    err := entity.UnmarshalJSON(data)
    require.Nil(t, err)

    assert.Equal(t, "FORMAT", entity.Format)
    assert.Equal(t, "CUR", entity.Currency)
    assert.Equal(t, true, entity.IsBalanceCheck)
  }

  t.Log("missing isBalanceCheck fallback to true")
  {
    data := []byte(`
      {
        "name": "NAME",
        "format": "FORMAT",
        "currency": "CUR"
      }
    `)
    entity := new(Account)
    err := entity.UnmarshalJSON(data)
    require.Nil(t, err)

    assert.Equal(t, "FORMAT", entity.Format)
    assert.Equal(t, "CUR", entity.Currency)
    assert.Equal(t, true, entity.IsBalanceCheck)
  }

  t.Log("name is required")
  {
    data := []byte(`
      {
        "format": "FORMAT",
        "currency": "CUR"
      }
    `)
    err := new(Account).UnmarshalJSON(data)
    require.NotNil(t, err)
  }

  t.Log("format is required")
  {
    data := []byte(`
      {
        "name": "NAME",
        "currency": "CUR"
      }
    `)
    err := new(Account).UnmarshalJSON(data)
    require.NotNil(t, err)
  }

  t.Log("currency is required")
  {
    data := []byte(`
      {
        "name": "NAME",
        "format": "FORMAT"
      }
    `)
    err := new(Account).UnmarshalJSON(data)
    require.NotNil(t, err)
  }

  t.Log("currency has charater limit")
  {
    lower := []byte(`
      {
        "name": "NAME",
        "format": "FORMAT",
        "currency": "CU"
      }
    `)
    require.NotNil(t, new(Account).UnmarshalJSON(lower))

    upper := []byte(`
      {
        "name": "NAME",
        "format": "FORMAT",
        "currency": "CURR"
      }
    `)
    require.NotNil(t, new(Account).UnmarshalJSON(upper))
  }
}
