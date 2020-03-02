package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalUnitConvert1(t *testing.T) {
	src := "10.12"
	res, err := UnitConverterRemovePoint(src)

	assert.NoError(t, err)
	assert.EqualValues(t, "10120000000000000000", res)
}

func TestNormalUnitConvert2(t *testing.T) {
	src := "1023"
	res, err := UnitConverterRemovePoint(src)

	assert.NoError(t, err)
	assert.Equal(t, "1023000000000000000000", res)
}

func TestFormatUnitConvert(t *testing.T) {
	src := "10.1.2"
	res, err := UnitConverterRemovePoint(src)

	assert.Error(t, err)
	assert.Equal(t, "0", res)
}

func TestLengthUnitConvert(t *testing.T) {
	src := "10.1234567890123456789"
	res, err := UnitConverterRemovePoint(src)

	assert.Error(t, err)
	assert.Equal(t, "0", res)
}

func TestRegexpUnitConvert1(t *testing.T) {
	src := "a10.23"
	res, err := UnitConverterRemovePoint(src)

	assert.Error(t, err)
	assert.Equal(t, "0", res)
}

func TestZeroUnitCounvert1(t *testing.T) {
	src := "0"
	res, err := UnitConverterRemovePoint(src)

	assert.NoError(t, err)
	assert.Equal(t, "0", res)
}

func TestZeroUnitCounvert2(t *testing.T) {
	src := "00.00"
	res, err := UnitConverterRemovePoint(src)

	assert.NoError(t, err)
	assert.Equal(t, "0", res)
}

func TestDecimalPlaceZeroUnitCounvert2(t *testing.T) {
	src := "0.01"
	res, err := UnitConverterRemovePoint(src)

	assert.NoError(t, err)
	assert.Equal(t, "10000000000000000", res)
}

func TestUnitConvertAddPoint1(t *testing.T) {
	src := "1"
	res := UnitConvertAddPoint(src)
	assert.Equal(t, "0.000000000000000001", res)
}

func TestUnitConvertAddPoint2(t *testing.T) {
	src := "1000000000000000001"
	res := UnitConvertAddPoint(src)
	assert.Equal(t, "1.000000000000000001", res)
}

func TestUnitConvertAddPoint3(t *testing.T) {
	src := "1123000000000000000"
	res := UnitConvertAddPoint(src)
	assert.Equal(t, "1.123", res)
}

func TestUnitConvertAddPoint4(t *testing.T) {
	src := "123000"
	res := UnitConvertAddPoint(src)
	assert.Equal(t, "0.000000000000123", res)
}

func TestUnitConvertAddPoint5(t *testing.T) {
	src := "0"
	res := UnitConvertAddPoint(src)
	assert.Equal(t, "0", res)
}

func TestUnitConvertAddPoint6(t *testing.T) {
	src := "10000000000000000000"
	res := UnitConvertAddPoint(src)
	assert.Equal(t, "10", res)
}
