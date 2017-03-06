package web

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestMockRequestBuilderWithPathf(t *testing.T) {
	assert := assert.New(t)
	rb := NewMockRequestBuilder(nil)
	rb.WithPathf("/test/%s", "foo")
	assert.Equal("/test/foo", rb.path)
}

func TestMockRequestBuilderWithVerb(t *testing.T) {
	assert := assert.New(t)
	rb := NewMockRequestBuilder(nil)
	rb.WithVerb("get")
	assert.Equal("GET", rb.verb)
}

func TestMockRequestBuilderWithQueryString(t *testing.T) {
	assert := assert.New(t)
	rb := NewMockRequestBuilder(nil)
	rb.WithQueryString("foo", "bar")
	assert.Equal("bar", rb.queryString.Get("foo"))
}

func TestMockRequestBuilderFetchResponseAsBytes(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/test_path", func(r *Ctx) Result {
		return r.Raw([]byte("test"))
	})
	resBody, err := app.Mock().WithPathf("/test_path").Bytes()
	assert.Nil(err)
	assert.NotEmpty(resBody)
	assert.Equal("test", string(resBody))
}

func TestMockRequestBuilderJSON(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/test_path", func(r *Ctx) Result {
		return r.RawJSON([]string{"foo", "bar"})
	})
	var res []string
	err := app.Mock().WithPathf("/test_path").JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res)
	assert.Equal("foo", res[0])
	assert.Equal("bar", res[1])
}
