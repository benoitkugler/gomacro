// Package echo is a substitute for the http framework echo package.
package echo

import (
	"iter"
	"mime/multipart"
	"net/http"
)

type Context interface {
	Response() *http.Response
	Bind(interface{}) error
	JSON(int, interface{}) error
	QueryParam(string) string
	Blob(code int, contentType string, b []byte) error
	FormValue(name string) string
	FormFile(name string) (*multipart.FileHeader, error)
}

type Echo struct{}

func (Echo) GET(string, func(Context) error, ...any) {}
func (Echo) POST(string, func(Context) error)        {}
func (Echo) PUT(string, func(Context) error)         {}
func (Echo) DELETE(string, func(Context) error)      {}

func StreamJSON[K any](*http.Response, iter.Seq2[K, error]) error { return nil }
