package web

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

type testObject struct {
	Str     string
	Ptr     *string
	Int     int
	Int64   int64
	UInt16  uint16
	UInt    uint
	UInt32  uint32
	UInt64  uint64
	Float32 float32
	Float64 float64
}

func TestSetValueByName(t *testing.T) {
	assert := assert.New(t)

	myObj := testObject{}

	err := setValueByName(myObj, "Str", "hello")
	assert.NotNil(err)

	err = setValueByName(&myObj, "Str", "hello")
	assert.Nil(err)
	assert.Equal("hello", myObj.Str)

	testString := "hello2"
	err = setValueByName(&myObj, "Str", &testString)
	assert.Nil(err)
	assert.Equal("hello2", myObj.Str)

	err = setValueByName(&myObj, "Int", 1234)
	assert.Nil(err)
	assert.Equal(1234, myObj.Int)

	err = setValueByName(&myObj, "Int", 123456.00)
	assert.Nil(err)
	assert.Equal(123456, myObj.Int)

	err = setValueByName(&myObj, "Int", "12345")
	assert.Nil(err)
	assert.Equal(12345, myObj.Int)

	err = setValueByName(&myObj, "Int", "hello")
	assert.NotNil(err)
}
