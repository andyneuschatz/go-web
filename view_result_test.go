package web

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestViewResultRender(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc, err := NewMockRequestBuilder(nil).WithResponseBuffer(buffer).Ctx(nil)
	assert.Nil(err)

	testView := template.New("testView")
	testView.Parse("{{.ViewModel.Text}}")
	viewCache := template.Must(testView, nil)

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   "testView",
		viewCache:  NewViewCacheWithTemplates(viewCache),
	}

	err = vr.Render(rc)
	assert.Nil(err)

	assert.NotZero(buffer.Len())
	assert.True(strings.Contains(buffer.String(), "bar"))
}

func TestViewResultRenderError(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc, err := NewMockRequestBuilder(nil).WithResponseBuffer(buffer).Ctx(nil)
	assert.Nil(err)

	testView := template.New("testView")
	testView.Parse("{{.ViewModel.Foo}}")
	viewCache := template.Must(testView, nil)

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   "testView",
		viewCache:  NewViewCacheWithTemplates(viewCache),
	}

	err = vr.Render(rc)
	assert.NotNil(err)
	assert.Zero(buffer.Len())
}

func TestViewResultRenderErrorTemplate(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc, err := NewMockRequestBuilder(nil).WithResponseBuffer(buffer).Ctx(nil)
	assert.Nil(err)

	views := template.New("main")

	errorTemplate := views.New(DefaultTemplateInternalServerError)
	_, err = errorTemplate.Parse("{{.}}")
	assert.Nil(err)

	testView := views.New("testView")
	_, err = testView.Parse("Foo: {{.ViewModel.Foo}}")
	assert.Nil(err)

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   "testView",
		viewCache:  NewViewCacheWithTemplates(views),
	}

	err = vr.Render(rc)
	assert.Nil(err)
	assert.NotZero(buffer.Len())
	assert.Equal(http.StatusInternalServerError, rc.Response.StatusCode())
}

func TestViewResultErrorNestedViews(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	rc, err := NewMockRequestBuilder(nil).WithResponseBuffer(buffer).Ctx(nil)
	assert.Nil(err)

	views := template.New("main")

	errorTemplate := views.New(DefaultTemplateInternalServerError)
	_, err = errorTemplate.Parse("{{.}}")
	assert.Nil(err)

	outerView := views.New("outerView")
	outerView.Parse("Inner: {{ template \"innerView\" . }}")

	testView := views.New("innerView")
	_, err = testView.Parse("Foo: {{.ViewModel.Foo}}")
	assert.Nil(err)

	vr := &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  testViewModel{Text: "bar"},
		Template:   "outerView",
		viewCache:  NewViewCacheWithTemplates(views),
	}

	err = vr.Render(rc)
	assert.Nil(err)
	assert.NotZero(buffer.Len())
	assert.Equal(http.StatusInternalServerError, rc.Response.StatusCode())
	// we should only show the error message.
	assert.False(strings.Contains(buffer.String(), "Inner:"), buffer.String())
}
