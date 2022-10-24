package iorpcbench

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/hexilee/iorpc"
)

var (
	Dispatcher      = iorpc.NewDispatcher()
	ServiceNoop     iorpc.Service
	ServiceReadData iorpc.Service
)

func init() {
	ServiceNoop, _ = Dispatcher.AddService("Noop", func(clientAddr string, request iorpc.Request) (response *iorpc.Response, err error) {
		return &iorpc.Response{}, nil
	})
	ServiceReadData, _ = Dispatcher.AddService(
		"ReadData",
		func(clientAddr string, request iorpc.Request) (*iorpc.Response, error) {
			request.Body.Close()
			size := uint64(128 * 1024)
			offset := uint64(0)

			if len(request.Headers) != 0 {
				if s, ok := request.Headers["Size"].(uint64); ok {
					size = s
				}

				if o, ok := request.Headers["Offset"].(uint64); ok {
					offset = o
				}
			}

			switch ret := filePool.Get().(type) {
			case error:
				return nil, ret
			case *File:
				stat, err := ret.Stat()
				if err != nil {
					return nil, err
				}
				fileSize := stat.Size()
				if offset >= uint64(fileSize) {
					return nil, errors.New("bad request, offset out of range")
				}

				if offset > 0 {
					ret.Seek(int64(offset), 0)
				}

				if offset+size > uint64(fileSize) {
					size = uint64(fileSize) - offset
				}
				ret.Limit = size
				return &iorpc.Response{
					Body: iorpc.Body{
						Size:   size,
						Reader: ret,
					},
				}, nil
			default:
				return nil, errors.New("unknown type in file pool")
			}
		},
	)
}

var filePool = sync.Pool{
	New: func() any {
		file, err := os.OpenFile("../data/tmp/bigdata", os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		return &File{File: file}
	},
}

type File struct {
	*os.File
	Limit uint64
}

func (f *File) Close() error {
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}
	f.Limit = 0
	filePool.Put(f)
	return nil
}

func (f *File) WriteTo(w io.Writer) (int64, error) {
	var reader io.Reader = f.File
	if f.Limit > 0 {
		reader = io.LimitReader(reader, int64(f.Limit))
	}
	return io.Copy(w, reader)
}

func ListenAndServe(addr string) error {
	// Start rpc server serving registered service.
	s := &iorpc.Server{
		// Accept clients on this TCP address.
		Addr: addr,

		// Echo handler - just return back the message we received from the client
		Handler: Dispatcher.HandlerFunc(),
	}
	s.CloseBody = true
	return s.Serve()
}
