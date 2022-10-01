package gorpcbench

import (
	"github.com/valyala/gorpc"
)

func NewClient(addr string, conns int) *gorpc.Client {
	// Start rpc client connected to the server.
	c := gorpc.NewTCPClient(addr)
	c.DisableCompression = true
	c.Conns = conns
	c.CloseBody = true
	c.Start()
	return c
}
