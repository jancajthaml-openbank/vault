package model

import (
  "fmt"
  "io"
  "math/big"
  "strings"
)

type Dec struct {
  unscaled big.Int
  scale    int32
}

var zeros = "0000000000000000000000000000000000000000000000000000000000000000"
var lzeros = int32(len(zeros))
var bigInt = [...]*big.Int{
  big.NewInt(0),
  big.NewInt(1),
  big.NewInt(2),
  big.NewInt(3),
  big.NewInt(4),
  big.NewInt(5),
  big.NewInt(6),
  big.NewInt(7),
  big.NewInt(8),
  big.NewInt(9),
  big.NewInt(10),
}

var exp10cache [64]big.Int = func() [64]big.Int {
  e10, e10i := [64]big.Int{}, bigInt[1]
  for i := range e10 {
    e10[i].Set(e10i)
    e10i = new(big.Int).Mul(e10i, bigInt[10])
  }
  return e10
}()

func NewDec(unscaled int64, scale int32) *Dec {
  return new(Dec).SetUnscaled(unscaled).SetScale(scale)
}

func (x *Dec) Unscaled() (u int64, ok bool) {
  u = x.unscaled.Int64()
  var i big.Int
  ok = i.SetInt64(u).Cmp(&x.unscaled) == 0
  return
}

func (x *Dec) UnscaledBig() *big.Int {
  return &x.unscaled
}

func (z *Dec) SetScale(scale int32) *Dec {
  z.scale = scale
  return z
}

func (z *Dec) SetUnscaled(unscaled int64) *Dec {
  z.unscaled.SetInt64(unscaled)
  return z
}

func (z *Dec) SetUnscaledBig(unscaled *big.Int) *Dec {
  z.unscaled.Set(unscaled)
  return z
}

func (z *Dec) Set(x *Dec) *Dec {
  if z != x {
    z.SetUnscaledBig(x.UnscaledBig())
    z.SetScale(x.scale)
  }
  return z
}

func (x *Dec) Sign() int {
  return x.UnscaledBig().Sign()
}

func (x *Dec) Cmp(y *Dec) int {
  xx, yy := upscale(x, y)
  return xx.UnscaledBig().Cmp(yy.UnscaledBig())
}

func (z *Dec) Add(x, y *Dec) *Dec {
  xx, yy := upscale(x, y)
  z.SetScale(xx.scale)
  z.UnscaledBig().Add(xx.UnscaledBig(), yy.UnscaledBig())
  return z
}

func (z *Dec) Sub(x, y *Dec) *Dec {
  xx, yy := upscale(x, y)
  z.SetScale(xx.scale)
  z.UnscaledBig().Sub(xx.UnscaledBig(), yy.UnscaledBig())
  return z
}

func upscale(a, b *Dec) (*Dec, *Dec) {
  if a.scale == b.scale {
    return a, b
  }
  if a.scale > b.scale {
    return a, b.rescale(a.scale)
  }
  return a.rescale(b.scale), b
}

func exp10(x int32) *big.Int {
  if int(x) < len(exp10cache) {
    return &exp10cache[int(x)]
  }
  return new(big.Int).Exp(bigInt[10], big.NewInt(int64(x)), nil)
}

func (x *Dec) rescale(newScale int32) *Dec {
  if x == nil {
    return x
  }
  shift := newScale - x.scale
  switch {
  case shift < 0:
    e := exp10(-shift)
    return new(Dec).SetUnscaledBig(new(big.Int).Quo(x.UnscaledBig(), e)).SetScale(newScale)
  case shift > 0:
    e := exp10(shift)
    return new(Dec).SetUnscaledBig(new(big.Int).Mul(x.UnscaledBig(), e)).SetScale(newScale)
  }
  return x
}

func (x *Dec) String() string {
  if x == nil {
    return "<nil>"
  }

  numbers := x.UnscaledBig().Text(10)

  if x.scale <= 0 {
    if x.scale != 0 && x.unscaled.Sign() != 0 {
      n := -x.scale
      for i := int32(0); i < n; i += lzeros {
        if n > i+lzeros {
          numbers += zeros
        } else {
          numbers += zeros[0:n-i]
        }
      }
    }
    return numbers
  }

  var negbit int32
  if x.unscaled.Sign() == -1 {
    negbit = 1
  }

  lens := int32(len(numbers))

  if lens-negbit > x.scale {
    return numbers[:lens-x.scale] + "." + numbers[lens-x.scale:]
  }

  var result string
  if negbit == 1 {
    result = "-0."
  } else {
    result = "0."
  }

  n := x.scale-lens+negbit
  for i := int32(0); i < n; i += lzeros {
    if n > i+lzeros {
      result += zeros
    } else {
      result += zeros[0:n-i]
    }
  }

  result += numbers[negbit:]

  return result
}

func (z *Dec) scan(r io.RuneScanner) (*Dec, error) {
  unscaled := make([]byte, 0, 256)
  dp, dg := -1, -1
loop:
  for {
    ch, _, err := r.ReadRune()
    if err == io.EOF {
      break loop
    }
    if err != nil {
      return nil, err
    }
    switch {
    case ch == '+' || ch == '-':
      if len(unscaled) > 0 || dp >= 0 {
        r.UnreadRune()
        break loop
      }
    case ch == '.':
      if dp >= 0 {
        r.UnreadRune()
        break loop
      }
      dp = len(unscaled)
      continue
    case ch >= '0' && ch <= '9':
      if dg == -1 {
        dg = len(unscaled)
      }
    default:
      r.UnreadRune()
      break loop
    }
    unscaled = append(unscaled, byte(ch))
  }
  if dg == -1 {
    return nil, fmt.Errorf("no digits read")
  }
  if dp >= 0 {
    z.SetScale(int32(len(unscaled) - dp))
  } else {
    z.SetScale(0)
  }
  _, ok := z.UnscaledBig().SetString(string(unscaled), 10)
  if !ok {
    return nil, fmt.Errorf("invalid decimal: %s", string(unscaled))
  }
  return z, nil
}

func (z *Dec) SetString(s string) (*Dec, bool) {
  r := strings.NewReader(s)
  _, err := z.scan(r)
  if err != nil {
    return nil, false
  }
  _, _, err = r.ReadRune()
  if err != io.EOF {
    return nil, false
  }
  return z, true
}
