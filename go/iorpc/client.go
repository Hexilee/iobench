package iorpcbench

import (
	"github.com/hexilee/iorpc"
)

func NewClient(addr string, conns int) *iorpc.Client {
	// Start rpc client connected to the server.
	c := iorpc.NewTCPClient(addr)
	c.DisableCompression = true
	c.Conns = conns
	c.CloseBody = true
	c.Start()
	return c
}
