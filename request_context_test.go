package web

import (
	"net/http"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestRequestContextState(t *testing.T) {
	assert := assert.New(t)

	context := NewRequestContext(nil, nil, nil)
	context.SetState("foo", "bar")
	assert.Equal("bar", context.State("foo"))
}

func TestRequestContextParamQuery(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithQueryString("foo", "bar").RequestContext(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestRequestContextParamHeader(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithHeader("foo", "bar").RequestContext(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestRequestContextParamForm(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithFormValue("foo", "bar").RequestContext(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestRequestContextParamCookie(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithCookie(&http.Cookie{Name: "foo", Value: "bar"}).RequestContext(nil)
	assert.Nil(err)
	assert.Equal("bar", context.Param("foo"))
}

func TestRequestContextPostBodyAsString(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithPostBody([]byte("test payload")).RequestContext(nil)
	assert.Nil(err)
	assert.Equal("test payload", context.PostBodyAsString())
}

func TestRequestContextPostBodyAsJSON(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithPostBody([]byte(`{"test":"test payload"}`)).RequestContext(nil)
	assert.Nil(err)

	var contents map[string]interface{}
	err = context.PostBodyAsJSON(&contents)
	assert.Nil(err)
	assert.Equal("test payload", contents["test"])
}

func TestRequestContextPostedFiles(t *testing.T) {
	assert := assert.New(t)
	context, err := NewMockRequestBuilder(nil).WithPostedFile(PostedFile{Key: "file", FileName: "test.txt", Contents: []byte("this is only a test")}).RequestContext(nil)
	assert.Nil(err)

	postedFiles, err := context.PostedFiles()
	assert.Nil(err)
	assert.NotEmpty(postedFiles)
	assert.Equal("file", postedFiles[0].Key)
	assert.Equal("test.txt", postedFiles[0].FileName)
	assert.Equal("this is only a test", string(postedFiles[0].Contents))
}

func TestRequestContextRouteParam(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).RequestContext(RouteParameters{"foo": "bar"})
	assert.Nil(err)
	value, err := context.RouteParam("foo")
	assert.Nil(err)
	assert.Equal("bar", value)
}

func TestRequestContextRouteParamInt(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).RequestContext(RouteParameters{"foo": "1"})
	assert.Nil(err)
	value, err := context.RouteParamInt("foo")
	assert.Nil(err)
	assert.Equal(1, value)
}

func TestRequestContextRouteParamInt64(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).RequestContext(RouteParameters{"foo": "1"})
	assert.Nil(err)
	value, err := context.RouteParamInt64("foo")
	assert.Nil(err)
	assert.Equal(1, value)
}

func TestRequestContextGetCookie(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).WithCookie(&http.Cookie{Name: "foo", Value: "bar"}).RequestContext(nil)
	assert.Nil(err)
	assert.Equal("bar", context.GetCookie("foo").Value)
}

func TestRequestContextHeaderParam(t *testing.T) {
	assert := assert.New(t)

	context, err := NewMockRequestBuilder(nil).RequestContext(nil)
	assert.Nil(err)
	value, err := context.HeaderParam("test")
	assert.NotNil(err)
	assert.Empty(value)

	context, err = NewMockRequestBuilder(nil).WithHeader("test", "foo").RequestContext(nil)
	assert.Nil(err)
	value, err = context.HeaderParam("test")
	assert.Nil(err)
	assert.Equal("foo", value)
}
