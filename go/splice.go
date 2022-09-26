package main

import (
	"log"
	"net"
	"os"
)

type SendFileServer struct {
	path string
}

func NewSendFileServer(path string) *SendFileServer {
	return &SendFileServer{path}
}

func (s *SendFileServer) ListenAndServe(addr string) error {
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

func (s *SendFileServer) handleConn(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	defer tcpConn.Close()
	for {
		file, err := os.OpenFile(s.path, os.O_RDONLY, 0)
		if err != nil {
			log.Println(err)
			break
		}
		_, err = tcpConn.ReadFrom(file)
		file.Close()
		if err != nil {
			log.Println(err)
			break
		}
	}
}
