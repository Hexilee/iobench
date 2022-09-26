package gorpcbench

import (
	"github.com/valyala/gorpc"
)

func NewClient(addr string) *gorpc.DispatcherClient {
	// Start rpc client connected to the server.
	c := gorpc.NewTCPClient(addr)
	c.Start()
	return newDispatcher().NewFuncClient(c)
}
