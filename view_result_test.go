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
	rc, err := NewMockRequestBuilder(nil).WithResponseBuffer(buffer).RequestContext(nil)
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
