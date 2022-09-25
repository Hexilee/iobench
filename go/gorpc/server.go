package gorpc

import (
	"os"

	"github.com/valyala/gorpc"
)

func newDispatcher() *gorpc.Dispatcher {
	d := gorpc.NewDispatcher()
	d.AddFunc("Get", func() ([]byte, error) {
		return os.ReadFile("../data/data")
	})
	return d
}

func ListenAndServe(addr string) error {
	// Start rpc server serving registered service.
	s := gorpc.NewTCPServer(addr, newDispatcher().NewHandlerFunc())
	return s.Serve()
}
