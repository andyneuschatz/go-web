package web

import (
	"net/http"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestCtxState(t *testing.T) {
	assert := assert.New(t)

	context := NewCtx(nil, nil, nil)
	context.SetState("foo", "bar")
	assert.Equal("bar", context.State("foo"))
}

func TestCtxParamQuery(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithQueryString("foo", "bar").Ctx(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestCtxParamHeader(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithHeader("foo", "bar").Ctx(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestCtxParamForm(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithFormValue("foo", "bar").Ctx(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestCtxParamCookie(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithCookie(&http.Cookie{Name: "foo", Value: "bar"}).Ctx(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestCtxPostBodyAsString(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithPostBody([]byte("test payload")).Ctx(nil)
	assert.Nil(err)
	body, err := context.PostBodyAsString()
	assert.Nil(err)
	assert.Equal("test payload", body)
}

func TestCtxPostBodyAsJSON(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithPostBody([]byte(`{"test":"test payload"}`)).Ctx(nil)
	assert.Nil(err)

	var contents map[string]interface{}
	err = context.PostBodyAsJSON(&contents)
	assert.Nil(err)
	assert.Equal("test payload", contents["test"])
}

func TestCtxPostedFiles(t *testing.T) {
	assert := assert.New(t)
	context, err := NewMockRequestBuilder(nil).WithPostedFile(PostedFile{Key: "file", FileName: "test.txt", Contents: []byte("this is only a test")}).Ctx(nil)
	assert.Nil(err)

	postedFiles, err := context.PostedFiles()
	assert.Nil(err)
	assert.NotEmpty(postedFiles)
	assert.Equal("file", postedFiles[0].Key)
	assert.Equal("test.txt", postedFiles[0].FileName)
	assert.Equal("this is only a test", string(postedFiles[0].Contents))
}

func TestCtxRouteParam(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).Ctx(RouteParameters{"foo": "bar"})
	assert.Nil(err)
	value, err := context.RouteParam("foo")
	assert.Nil(err)
	assert.Equal("bar", value)
}

func TestCtxRouteParamInt(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).Ctx(RouteParameters{"foo": "1"})
	assert.Nil(err)
	value, err := context.RouteParamInt("foo")
	assert.Nil(err)
	assert.Equal(1, value)
}

func TestCtxRouteParamInt64(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).Ctx(RouteParameters{"foo": "1"})
	assert.Nil(err)
	value, err := context.RouteParamInt64("foo")
	assert.Nil(err)
	assert.Equal(1, value)
}

func TestCtxGetCookie(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithCookie(&http.Cookie{Name: "foo", Value: "bar"}).Ctx(nil)
	assert.Nil(err)
	assert.Equal("bar", context.GetCookie("foo").Value)
}

func TestCtxHeaderParam(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).Ctx(nil)
	assert.Nil(err)
	value, err := context.HeaderParam("test")
	assert.NotNil(err)
	assert.Empty(value)

	context, err = NewMockRequestBuilder(nil).WithHeader("test", "foo").Ctx(nil)
	assert.Nil(err)
	value, err = context.HeaderParam("test")
	assert.Nil(err)
	assert.Equal("foo", value)
}

func TestCtxWriteNewCookie(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).Ctx(nil)
	assert.Nil(err)

	context.WriteNewCookie("foo", "bar", nil, "/foo/bar", true)
	assert.Equal("foo=bar; Path=/foo/bar; HttpOnly; Secure", context.Response.Header().Get("Set-Cookie"))
}
