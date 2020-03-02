package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalUnitConvert1(t *testing.T) {
	src := Hdac("10.12")
	res, err := ToBigsun(src)

	assert.NoError(t, err)
	assert.EqualValues(t, Bigsun("10120000000000000000"), res)
}

func TestNormalUnitConvert2(t *testing.T) {
	src := Hdac("1023")
	res, err := ToBigsun(src)

	assert.NoError(t, err)
	assert.Equal(t, Bigsun("1023000000000000000000"), res)
}

func TestFormatUnitConvert(t *testing.T) {
	src := Hdac("10.1.2")
	res, err := ToBigsun(src)

	assert.Error(t, err)
	assert.Equal(t, Bigsun("0"), res)
}

func TestLengthUnitConvert(t *testing.T) {
	src := Hdac("10.1234567890123456789")
	res, err := ToBigsun(src)

	assert.Error(t, err)
	assert.Equal(t, Bigsun("0"), res)
}

func TestRegexpUnitConvert1(t *testing.T) {
	src := Hdac("a10.23")
	res, err := ToBigsun(src)

	assert.Error(t, err)
	assert.Equal(t, Bigsun("0"), res)
}

func TestZeroUnitCounvert1(t *testing.T) {
	src := Hdac("0")
	res, err := ToBigsun(src)

	assert.NoError(t, err)
	assert.Equal(t, Bigsun("0"), res)
}

func TestZeroUnitCounvert2(t *testing.T) {
	src := Hdac("00.00")
	res, err := ToBigsun(src)

	assert.NoError(t, err)
	assert.Equal(t, Bigsun("0"), res)
}

func TestDecimalPlaceZeroUnitCounvert2(t *testing.T) {
	src := Hdac("0.01")
	res, err := ToBigsun(src)

	assert.NoError(t, err)
	assert.Equal(t, Bigsun("10000000000000000"), res)
}

func TestUnitConvertAddPoint1(t *testing.T) {
	src := Bigsun("1")
	res := ToHdac(src)
	assert.Equal(t, Hdac("0.000000000000000001"), res)
}

func TestUnitConvertAddPoint2(t *testing.T) {
	src := Bigsun("1000000000000000001")
	res := ToHdac(src)
	assert.Equal(t, Hdac("1.000000000000000001"), res)
}

func TestUnitConvertAddPoint3(t *testing.T) {
	src := Bigsun("1123000000000000000")
	res := ToHdac(src)
	assert.Equal(t, Hdac("1.123"), res)
}

func TestUnitConvertAddPoint4(t *testing.T) {
	src := Bigsun("123000")
	res := ToHdac(src)
	assert.Equal(t, Hdac("0.000000000000123"), res)
}

func TestUnitConvertAddPoint5(t *testing.T) {
	src := Bigsun("0")
	res := ToHdac(src)
	assert.Equal(t, Hdac("0"), res)
}

func TestUnitConvertAddPoint6(t *testing.T) {
	src := Bigsun("10000000000000000000")
	res := ToHdac(src)
	assert.Equal(t, Hdac("10"), res)
}
