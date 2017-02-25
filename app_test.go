package web

import (
	"bytes"
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-logger"
	_ "github.com/lib/pq"
)

func TestAppNoDiagnostics(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.GET("/", func(c *Ctx) Result {
		return c.Raw([]byte("ok!"))
	})

	assert.Nil(app.Mock().Get("/").Execute())
}

func TestAppSetDiagnostics(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.SetLogger(logger.New(logger.NewEventFlagSetAll()))
	assert.NotNil(app.Logger())
	assert.True(app.Logger().Events().IsAllEnabled())
}

func TestAppInitializeConfig(t *testing.T) {
	assert := assert.New(t)

	app := New()
	err := app.InitializeConfig(&myConfig{})
	assert.Nil(err)
	assert.Equal("8080", app.Config().(*myConfig).Port)
}

func TestAppCtx(t *testing.T) {
	assert := assert.New(t)

	response := bytes.NewBuffer([]byte{})

	app := New()

	rc, err := app.Mock().WithResponseBuffer(response).Ctx(nil)
	assert.Nil(err)
	assert.NotNil(rc)
	assert.Nil(rc.logger)

	result := rc.Raw([]byte("foo"))
	assert.NotNil(result)

	err = result.Render(rc)
	assert.Nil(err)
	assert.NotZero(response.Len())
	assert.True(strings.Contains(response.String(), "foo"))
}

func TestAppStaticRewrite(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.AddStaticRewriteRule("/testPath/*filepath", "(.*)", func(path string, pieces ...string) string {
		return path
	})

	assert.NotEmpty(app.staticRewriteRules)
}

func TestAppStaticRewriteBadExp(t *testing.T) {
	assert := assert.New(t)
	app := New()
	err := app.AddStaticRewriteRule("/testPath/*filepath", "((((", func(path string, pieces ...string) string {
		return path
	})

	assert.NotNil(err)
	assert.Empty(app.staticRewriteRules)
}

func TestAppStaticHeader(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.AddStaticHeader("/testPath/*filePath", "cache-control", "haha what is caching.")
	assert.NotEmpty(app.staticHeaders)
	assert.NotEmpty(app.staticHeaders["/testPath/*filePath"])
}

func TestAppMiddleWarePipeline(t *testing.T) {
	assert := assert.New(t)
	app := New()

	didRun := false
	app.GET("/",
		func(r *Ctx) Result { return r.Raw([]byte("OK!")) },
		func(action Action) Action {
			didRun = true
			return action
		},
		func(action Action) Action {
			return func(r *Ctx) Result {
				return r.Raw([]byte("foo"))
			}
		},
	)

	result, err := app.Mock().WithPathf("/").FetchResponseAsBytes()
	assert.Nil(err)
	assert.True(didRun)
	assert.Equal("foo", string(result))
}

func TestAppMockTransactions(t *testing.T) {
	assert := assert.New(t)
	app := New()

	tx := &sql.Tx{}
	app.IsolateTo(tx)

	var action = func(r *Ctx) Result {
		assert.NotNil(r.Tx())
		return r.Raw([]byte("OK!"))
	}

	app.GET("/", action)

	err := app.Mock().WithPathf("/").Execute()
	assert.Nil(err)
}

func TestAppStatic(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.Static("/static/*filepath", http.Dir("testdata"))

	index, err := app.Mock().WithPathf("/static/test_file.html").FetchResponseAsBytes()
	assert.Nil(err)
	assert.True(strings.Contains(string(index), "Test!"), string(index))
}

func TestAppStaticSingleFile(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/", func(r *Ctx) Result {
		return r.Static("testdata/test_file.html")
	})

	index, err := app.Mock().WithPathf("/").FetchResponseAsBytes()
	assert.Nil(err)
	assert.True(strings.Contains(string(index), "Test!"), string(index))
}

func TestAppProviderMiddleware(t *testing.T) {
	assert := assert.New(t)

	var okAction = func(r *Ctx) Result {
		assert.NotNil(r.DefaultResultProvider())
		return r.Raw([]byte("OK!"))
	}

	app := New()
	app.GET("/", okAction, APIProviderAsDefault)

	err := app.Mock().WithPathf("/").Execute()
	assert.Nil(err)
}

func TestAppProviderMiddlewareOrder(t *testing.T) {
	assert := assert.New(t)

	app := New()

	var okAction = func(r *Ctx) Result {
		assert.NotNil(r.DefaultResultProvider())
		return r.Raw([]byte("OK!"))
	}

	var dependsOnProvider = func(action Action) Action {
		return func(r *Ctx) Result {
			assert.NotNil(r.DefaultResultProvider())
			return action(r)
		}
	}

	app.GET("/", okAction, dependsOnProvider, APIProviderAsDefault)

	err := app.Mock().WithPathf("/").Execute()
	assert.Nil(err)
}

func TestAppDefaultResultProvider(t *testing.T) {
	assert := assert.New(t)
	app := New()
	assert.NotNil(app.DefaultResultProvider())

	rc := app.newCtx(nil, nil, nil)
	assert.Nil(rc.view)
	assert.NotNil(rc.text, "rc.text should be provided as default")
	assert.NotNil(rc.defaultResultProvider)
}

func TestAppDefaultResultProviderWithDefault(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.SetDefaultResultProvider(ViewProviderAsDefault)
	assert.NotNil(app.DefaultResultProvider())

	rc := app.newCtx(nil, nil, nil)
	assert.NotNil(rc.view)
	assert.Nil(rc.api)
	assert.NotNil(rc.defaultResultProvider)
}

func TestAppDefaultResultProviderWithDefaultFromRoute(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.View().Templates().New(DefaultTemplateNotAuthorized).Parse("<html><body><h4>Not Authorized</h4></body></html>")
	app.SetDefaultResultProvider(APIProviderAsDefault)
	app.GET("/", controllerNoOp, SessionRequired, ViewProviderAsDefault)

	//somehow assert that the content type is html
	meta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(ContentTypeHTML, meta.Headers.Get("Content-Type"))
}

func TestAppViewResult(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.View().AddPaths("testdata/test_file.html")
	app.GET("/", func(r *Ctx) Result {
		return r.View().View("test", "foobarbaz")
	})

	meta, res, err := app.Mock().WithPathf("/").FetchResponseAsBytesWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.Equal(ContentTypeHTML, meta.Headers.Get(HeaderContentType))
	assert.Contains("foobarbaz", string(res))
}
