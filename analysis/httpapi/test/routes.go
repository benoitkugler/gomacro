package main

import (
	"fmt"
	"iter"

	"github.com/benoitkugler/gomacro/analysis/httpapi/test/echo"
	"github.com/benoitkugler/gomacro/analysis/httpapi/test/inner"
)

const route = "/const_url_from_package/"

type IdDossier int64

type controller struct{}

func QueryParamInt[T ~int64](echo.Context, string) (T, error) { return 0, nil }
func (controller) QueryParamInt64(echo.Context, string) int64 { return 0 }
func (controller) QueryParamBool(echo.Context, string) bool   { return false }
func (controller) JWTMiddlewareForQuery() bool                { return false }
func FormValueJSON(echo.Context, string, any) error           { return nil }

func (controller) handle1(c echo.Context) error {
	var (
		in  int
		out string
	)
	if err := c.Bind(&in); err != nil {
		return err
	}
	return c.JSON(200, out)
}

func handler(echo.Context) error { return nil }

func (controller) handler2(c echo.Context) error {
	return c.JSON(200, controller{})
}

func (controller) handler3(echo.Context) error { return nil }
func (controller) handler4(echo.Context) error { return nil }
func (controller) handler5(echo.Context) error { return nil }
func (controller) handler6(echo.Context) error { return nil }

// special converters
func (ct controller) handler7(c echo.Context) error {
	var v uint32
	p1 := ct.QueryParamBool(c, "my-bool")
	p2 := ct.QueryParamInt64(c, "my-int")
	_ = FormValueJSON(c, "json-field", &v)
	fmt.Println(p1, p2)
	var code uint
	return c.JSON(200, code)
}

func (controller) handler8(c echo.Context) error {
	id1, id2 := c.QueryParam("query_param1"), c.QueryParam("query_param2")
	v1 := c.FormValue("value_1")
	v2, _ := c.FormFile("file_2")

	fmt.Println(id1, id2, v1, v2)
	var code uint
	return c.JSON(200, code)
}

// with Blob
func (controller) handler9(c echo.Context) error {
	var output []byte
	return c.Blob(200, "", output)
}

// with Generic
func (controller) handler10(c echo.Context) error {
	v, err := QueryParamInt[IdDossier](c, "param-name")
	if err != nil {
		return err
	}
	_ = v
	var output []byte
	return c.Blob(200, "", output)
}

// with stream
func (ct controller) handler11(c echo.Context) error {
	var it iter.Seq2[int, error]
	return echo.StreamJSON(c.Response(), it)
}

func routes(e *echo.Echo, ct *controller, ct2 inner.Controller) {
	e.GET(route, handler)
	const routeFunc = "const_local_url"
	e.GET(routeFunc, ct.handle1)
	e.POST(inner.Url, ct2.HandleExt)
	e.POST(inner.Url+"endpoint", ct.handler2)
	e.POST(inner.Url+"endpoint/"+"entoher/"+routeFunc, func(ctx echo.Context) error { return nil })
	e.POST("host"+inner.Url, ct.handler3)
	e.POST("host"+"endpoint", ct.handler4)
	e.POST("/string_litteral", ct.handler5)
	e.PUT("/with_param/:param", ct.handler6)
	e.DELETE("/special_param_value/:class/route", ct.handler7)
	e.DELETE("/special_param_value/:default/route", ct.handler8)

	e.GET("/extern function", inner.TopLevel)

	e.GET("/func litteral", func(ctx echo.Context) error {
		return ctx.JSON(200, [][]string{{"a"}, {"b"}})
	})

	e.POST("/download", ct.handler9)
	e.DELETE("/with_generic", ct.handler10)
	e.GET("/with middleware", func(ctx echo.Context) error { return nil }, ct.JWTMiddlewareForQuery())

	e.GET("/with_json_stream", ct.handler11)

	e.GET("/ignore", ct.handle1) // ignore
}
