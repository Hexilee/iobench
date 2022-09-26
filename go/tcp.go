package main

import (
	"io"
	"log"
	"net"
	"os"
)

type TCPFileServer struct {
	path string
}

func NewTCPFileServer(path string) *TCPFileServer {
	return &TCPFileServer{path}
}

func (s *TCPFileServer) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *TCPFileServer) handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		file, err := os.OpenFile(s.path, os.O_RDONLY, 0)
		if err != nil {
			log.Println(err)
			break
		}
		_, err = io.Copy(conn, file)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
