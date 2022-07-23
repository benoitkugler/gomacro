// Package echo is a substitute for the http framework echo package.
package echo

type Context interface {
	Bind(interface{}) error
	JSON(int, interface{}) error
	QueryParam(string) string
}

type Echo struct{}

func (Echo) GET(string, func(Context) error)    {}
func (Echo) POST(string, func(Context) error)   {}
func (Echo) PUT(string, func(Context) error)    {}
func (Echo) DELETE(string, func(Context) error) {}
