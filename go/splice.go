package main

import (
	"log"
	"net"
	"os"
	"syscall"

	"github.com/hanwen/go-fuse/splice"
)

const (
	spliceOffset = 1
	spliceSize   = 128 * 1024
)

type SpliceServer struct {
	file *os.File
}

func NewSpliceServer(path string) (*SpliceServer, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return &SpliceServer{file}, nil
}

func (s *SpliceServer) ListenAndServe(addr string) error {
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

func (s *SpliceServer) handleConn(conn net.Conn) {
	defer conn.Close()
	tcpConn, err := conn.(syscall.Conn).SyscallConn()
	if err != nil {
		log.Printf("getting syscall conn from tcp conn failed: %v", err)
		return
	}
	pair, err := splice.Get()
	if err != nil {
		log.Printf("splice.Get() failed: %v", err)
		return
	}
	defer splice.Done(pair)
	if err := pair.Grow(spliceSize); err != nil {
		log.Printf("pair.Grow() failed: %v", err)
		return
	}

	for {
		written := 0
		_, err := pair.LoadFromAt(s.file.Fd(), spliceSize, spliceOffset)
		if err != nil {
			log.Printf("pair.LoadFromAt() failed: %v", err)
			break
		}
		var writeError error
		err = tcpConn.Write(func(fd uintptr) (done bool) {
			var n int
			n, writeError = pair.WriteTo(fd, spliceSize-written)
			if err != nil {
				log.Printf("pair.WriteTo() failed: %v", err)
				return true
			}
			written += n
			return written == spliceSize
		})
		if err != nil {
			log.Println(writeError)
			break
		}
		if writeError != nil {
			log.Println(writeError)
			break
		}
	}
}
