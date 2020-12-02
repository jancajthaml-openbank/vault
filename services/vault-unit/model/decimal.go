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

var zeros = []byte("0000000000000000000000000000000000000000000000000000000000000000")

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

// NewDec allocates and returns a new Dec set to the given int64 unscaled value
// and scale.
func NewDec(unscaled int64, scale int32) *Dec {
  return new(Dec).SetUnscaled(unscaled).SetScale(scale)
}

// Scale returns the scale of x.
func (x *Dec) Scale() int32 {
  return x.scale
}

// Unscaled returns the unscaled value of x for u and true for ok when the
// unscaled value can be represented as int64; otherwise it returns an undefined
// int64 value for u and false for ok. Use x.UnscaledBig().Int64() to avoid
// checking the validity of the value when the check is known to be redundant.
func (x *Dec) Unscaled() (u int64, ok bool) {
  u = x.unscaled.Int64()
  var i big.Int
  ok = i.SetInt64(u).Cmp(&x.unscaled) == 0
  return
}

// UnscaledBig returns the unscaled value of x as *big.Int.
func (x *Dec) UnscaledBig() *big.Int {
  return &x.unscaled
}

// SetScale sets the scale of z, with the unscaled value unchanged, and returns
// z.
// The mathematical value of the Dec changes as if it was multiplied by
// 10**(oldscale-scale).
func (z *Dec) SetScale(scale int32) *Dec {
  z.scale = scale
  return z
}

// SetUnscaled sets the unscaled value of z, with the scale unchanged, and
// returns z.
func (z *Dec) SetUnscaled(unscaled int64) *Dec {
  z.unscaled.SetInt64(unscaled)
  return z
}

// SetUnscaledBig sets the unscaled value of z, with the scale unchanged, and
// returns z.
func (z *Dec) SetUnscaledBig(unscaled *big.Int) *Dec {
  z.unscaled.Set(unscaled)
  return z
}

// Set sets z to the value of x and returns z.
// It does nothing if z == x.
func (z *Dec) Set(x *Dec) *Dec {
  if z != x {
    z.SetUnscaledBig(x.UnscaledBig())
    z.SetScale(x.Scale())
  }
  return z
}

// Sign returns:
//
//  -1 if x <  0
//   0 if x == 0
//  +1 if x >  0
//
func (x *Dec) Sign() int {
  return x.UnscaledBig().Sign()
}

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
func (x *Dec) Cmp(y *Dec) int {
  xx, yy := upscale(x, y)
  return xx.UnscaledBig().Cmp(yy.UnscaledBig())
}

// Add sets z to the sum x+y and returns z.
// The scale of z is the greater of the scales of x and y.
func (z *Dec) Add(x, y *Dec) *Dec {
  xx, yy := upscale(x, y)
  z.SetScale(xx.Scale())
  z.UnscaledBig().Add(xx.UnscaledBig(), yy.UnscaledBig())
  return z
}

// Sub sets z to the difference x-y and returns z.
// The scale of z is the greater of the scales of x and y.
func (z *Dec) Sub(x, y *Dec) *Dec {
  xx, yy := upscale(x, y)
  z.SetScale(xx.Scale())
  z.UnscaledBig().Sub(xx.UnscaledBig(), yy.UnscaledBig())
  return z
}

func upscale(a, b *Dec) (*Dec, *Dec) {
  if a.Scale() == b.Scale() {
    return a, b
  }
  if a.Scale() > b.Scale() {
    return a, b.rescale(a.Scale())
  }
  return a.rescale(b.Scale()), b
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
  shift := newScale - x.Scale()
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

func appendZeros(s []byte, n int32) []byte {
  for i := int32(0); i < n; i += lzeros {
    if n > i+lzeros {
      s = append(s, zeros...)
    } else {
      s = append(s, zeros[0:n-i]...)
    }
  }
  return s
}

func (x *Dec) String() string {
  if x == nil {
    return "<nil>"
  }
  scale := x.Scale()
  s := []byte(x.UnscaledBig().String())
  if scale <= 0 {
    if scale != 0 && x.unscaled.Sign() != 0 {
      s = appendZeros(s, -scale)
    }
    return string(s)
  }
  negbit := int32(-((x.Sign() - 1) / 2))
  // scale > 0
  lens := int32(len(s))
  if lens-negbit <= scale {
    ss := make([]byte, 0, scale+2)
    if negbit == 1 {
      ss = append(ss, '-')
    }
    ss = append(ss, '0', '.')
    ss = appendZeros(ss, scale-lens+negbit)
    ss = append(ss, s[negbit:]...)
    return string(ss)
  }
  // lens > scale
  ss := make([]byte, 0, lens+1)
  ss = append(ss, s[:lens-scale]...)
  ss = append(ss, '.')
  ss = append(ss, s[lens-scale:]...)
  return string(ss)
}

func (z *Dec) scan(r io.RuneScanner) (*Dec, error) {
  unscaled := make([]byte, 0, 256) // collects chars of unscaled as bytes
  dp, dg := -1, -1                 // indexes of decimal point, first digit
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
      if len(unscaled) > 0 || dp >= 0 { // must be first character
        r.UnreadRune()
        break loop
      }
    case ch == '.':
      if dp >= 0 {
        r.UnreadRune()
        break loop
      }
      dp = len(unscaled)
      continue // don't add to unscaled
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

// SetString sets z to the value of s, interpreted as a decimal (base 10),
// and returns z and a boolean indicating success. The scale of z is the
// number of digits after the decimal point (including any trailing 0s),
// or 0 if there is no decimal point. If SetString fails, the value of z
// is undefined but the returned value is nil.
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
