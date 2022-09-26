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
	chunk := make([]byte, 128<<10)
	for {
		isBreak := func() bool {
			file, err := os.OpenFile(s.path, os.O_RDONLY, 0)
			if err != nil {
				log.Println(err)
				return true
			}
			defer file.Close()
			for {
				n, err := file.Read(chunk)
				if err != nil {
					if err != io.EOF {
						log.Println(err)
					}
					return false
				}
				if _, err := conn.Write(chunk[:n]); err != nil {
					log.Println(err)
					return true
				}
			}
		}()
		if isBreak {
			break
		}
	}
}
