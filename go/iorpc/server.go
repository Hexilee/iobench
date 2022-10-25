package iorpcbench

import (
	"errors"
	"os"

	"github.com/hexilee/iorpc"
)

var (
	Dispatcher      = iorpc.NewDispatcher()
	ServiceNoop     iorpc.Service
	ServiceReadData iorpc.Service
)

var (
	dataFile *os.File
	fileSize int64
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

			if offset >= uint64(fileSize) {
				return nil, errors.New("bad request, offset out of range")
			}

			if offset+size > uint64(fileSize) {
				size = uint64(fileSize) - offset
			}

			return &iorpc.Response{
				Body: iorpc.Body{
					Offset:   offset,
					Size:     size,
					Reader:   &File{file: dataFile},
					NotClose: true,
				},
			}, nil
		},
	)
}

type File struct {
	file *os.File
}

func (f *File) Close() error {
	return f.file.Close()
}

func (f *File) File() *os.File {
	return f.file
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func ListenAndServe(addr string) error {
	file, err := os.OpenFile("../data/tmp/bigdata", os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	dataFile, fileSize = file, stat.Size()

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
