package web

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-logger"
	_ "github.com/lib/pq"
)

func controllerNoOp(_ *Ctx) Result { return nil }

func TestAppNoDiagnostics(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	app := New()
	app.GET("/", func(c *Ctx) Result {
		route = c.Route()
		return c.Raw([]byte("ok!"))
	})

	assert.Nil(app.Mock().Get("/").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/", route.Path)
	assert.NotNil(route.Handler)
}

func TestAppPathParams(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	var params RouteParameters
	app := New()
	app.GET("/:uuid", func(c *Ctx) Result {
		route = c.Route()
		params = c.routeParameters
		return c.Raw([]byte("ok!"))
	})

	assert.Nil(app.Mock().Get("/foo").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/:uuid", route.Path)
	assert.NotNil(route.Handler)

	assert.NotNil(params)
	assert.NotEmpty(params)
	assert.Equal("foo", params.Get("uuid"))
}

func TestAppPathParamsForked(t *testing.T) {
	assert := assert.New(t)

	var route *Route
	var params RouteParameters
	app := New()
	app.GET("/foos/bar/:uuid", func(c *Ctx) Result {
		route = c.Route()
		params = c.routeParameters
		return c.Raw([]byte("ok!"))
	})
	app.GET("/foo/:uuid", func(c *Ctx) Result { return nil })

	assert.Nil(app.Mock().Get("/foos/bar/foo").Execute())
	assert.NotNil(route)
	assert.Equal("GET", route.Method)
	assert.Equal("/foos/bar/:uuid", route.Path)
	assert.NotNil(route.Handler)

	assert.NotNil(params)
	assert.NotEmpty(params)
	assert.Equal("foo", params.Get("uuid"))
}

func TestAppSetDiagnostics(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.SetLogger(logger.New(logger.NewEventFlagSetAll()))
	assert.NotNil(app.Logger())
	assert.True(app.Logger().Events().IsAllEnabled())
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

	result, err := app.Mock().WithPathf("/").Bytes()
	assert.Nil(err)
	assert.True(didRun)
	assert.Equal("foo", string(result))
}

func TestAppStatic(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.Static("/static/*filepath", http.Dir("testdata"))

	index, err := app.Mock().WithPathf("/static/test_file.html").Bytes()
	assert.Nil(err)
	assert.True(strings.Contains(string(index), "Test!"), string(index))
}

func TestAppStaticSingleFile(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/", func(r *Ctx) Result {
		return r.Static("testdata/test_file.html")
	})

	index, err := app.Mock().WithPathf("/").Bytes()
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
	assert.Nil(app.DefaultMiddleware())

	rc := app.newCtx(nil, nil, nil, nil)
	assert.Nil(rc.view)
	assert.NotNil(rc.text, "rc.text should be provided as default")
	assert.NotNil(rc.defaultResultProvider)
}

func TestAppDefaultResultProviderWithDefault(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.SetDefaultMiddleware(ViewProviderAsDefault)
	assert.NotNil(app.DefaultMiddleware())

	rc := app.newCtx(nil, nil, nil, nil)
	assert.Nil(rc.view)
	assert.Nil(rc.api)

	// this will be set to the default initially
	assert.NotNil(rc.defaultResultProvider)

	app.GET("/", func(ctx *Ctx) Result {
		assert.NotNil(ctx.DefaultResultProvider())
		_, isTyped := ctx.DefaultResultProvider().(*ViewResultProvider)
		assert.True(isTyped)
		return nil
	})
}

func TestAppDefaultResultProviderWithDefaultFromRoute(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.ViewCache().Templates().New(DefaultTemplateNotAuthorized).Parse("<html><body><h4>Not Authorized</h4></body></html>")
	app.SetDefaultMiddleware(APIProviderAsDefault)
	app.GET("/", controllerNoOp, SessionRequired, ViewProviderAsDefault)

	//somehow assert that the content type is html
	meta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(ContentTypeHTML, meta.Headers.Get(HeaderContentType))
}

func TestAppViewResult(t *testing.T) {
	assert := assert.New(t)

	app := New()
	app.ViewCache().AddPaths("testdata/test_file.html")
	app.GET("/", func(r *Ctx) Result {
		return r.View().View("test", "foobarbaz")
	})

	res, meta, err := app.Mock().WithPathf("/").BytesWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.Equal(ContentTypeHTML, meta.Headers.Get(HeaderContentType))
	assert.Contains("foobarbaz", string(res))
}
