package model

import (
  "testing"

  "github.com/stretchr/testify/assert"
)

func Test_String(t *testing.T) {

  t.Log("zero value")
  {
    assert.Equal(t, "0", new(Dec).String())
  }

  t.Log("-1.0")
  {
    entity, _ := new(Dec).SetString("-1.0")
    assert.Equal(t, "-1.0", entity.String())
  }

  t.Log("1.0")
  {
    entity, _ := new(Dec).SetString("1.0")
    assert.Equal(t, "1.0", entity.String())
  }

  t.Log("0.1")
  {
    entity, _ := new(Dec).SetString("0.1")
    assert.Equal(t, "0.1", entity.String())
  }

  t.Log("-0.1")
  {
    entity, _ := new(Dec).SetString("-0.1")
    assert.Equal(t, "-0.1", entity.String())
  }


  t.Log("10000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000001")
  {
    entity, _ := new(Dec).SetString("10000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000001")
    assert.Equal(t, "10000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000001", entity.String())
  }
}


func Test_Add(t *testing.T) {


  t.Log("10000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000001")
  {
    a, _ := new(Dec).SetString("10000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000002")
    b, _ := new(Dec).SetString("20000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000001")
    assert.Equal(t, "30000000000000000000000000000000000000000000000000000000.000000000000000000000000000000000000000000000000000000003", new(Dec).Add(a, b).String())
  }
}

func BenchmarkDec_String(b *testing.B) {
  entity := new(Dec)

  b.ReportAllocs()
  b.ResetTimer()
  for n := 0; n < b.N; n++ {
    entity.String()
  }
}
