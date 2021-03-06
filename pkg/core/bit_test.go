package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bit(t *testing.T) {
	a := assert.New(t)

	b := Bit(1)

	b.MarshalText()
	b.UnmarshalText([]byte(`1`))

	bBs, err := b.MarshalJSON()
	a.Nil(err)
	a.Equal(`1`, string(bBs))

	b = Bit(0)
	bBs, err = b.MarshalJSON()
	a.Nil(err)
	a.Equal(`0`, string(bBs))

	a.Nil(b.UnmarshalJSON([]byte(`1`)))
	a.True(b.Bool())

	a.Nil(b.UnmarshalJSON([]byte(`0`)))
	a.False(b.Bool())

	_, err = Bit(2).MarshalJSON()
	a.Contains(err.Error(), `invalid value 2, Bit only accepts 0 or 1`)
	a.Contains(b.UnmarshalJSON([]byte(`2`)).Error(), `invalid value 2, Bit only accepts 0 or 1`)

	bs := Bits{0, 1, 1, 0}

	bs.MarshalText()
	bs.UnmarshalText([]byte(`0101`))

	bsBs, err := bs.MarshalText()
	a.Nil(err)
	a.Equal(`0101`, string(bsBs))

	a.Nil(bs.UnmarshalText([]byte(`0110`)))
	_, err = Bits{0, 1, 2, 1, 0}.MarshalText()
	a.Contains(err.Error(), `invalid value 2, Bits only accepts 0s and 1s`)
	a.Contains(bs.UnmarshalText([]byte(`01210`)).Error(), `invalid value 2, Bits only accepts 0s and 1s`)
}
