package iorpcbench

import (
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/hexilee/iorpc"
)

var (
	Dispatcher        = iorpc.NewDispatcher()
	ServiceNoop       iorpc.Service
	ServiceReadData   iorpc.Service
	ServiceReadMemory iorpc.Service
)

type StaticBuffer []byte
type ReadHeaders struct {
	Offset, Size uint64
}

func (h *ReadHeaders) Encode(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, *h)
}

func (h *ReadHeaders) Decode(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, h)
}

var (
	dataFile *os.File
	fileSize int64

	staticData = make(StaticBuffer, 128*1024)
)

func init() {
	iorpc.RegisterHeaders(func() iorpc.Headers {
		return new(ReadHeaders)
	})

	ServiceNoop, _ = Dispatcher.AddService("Noop", func(clientAddr string, request iorpc.Request) (response *iorpc.Response, err error) {
		return &iorpc.Response{}, nil
	})
	ServiceReadData, _ = Dispatcher.AddService(
		"ReadData",
		func(clientAddr string, request iorpc.Request) (*iorpc.Response, error) {
			request.Body.Close()
			size := uint64(128 * 1024)
			offset := uint64(1)

			if request.Headers != nil {
				if headers := request.Headers.(*ReadHeaders); headers != nil {
					size = headers.Size
					offset = headers.Offset
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

	ServiceReadMemory, _ = Dispatcher.AddService(
		"ReadMemory",
		func(clientAddr string, request iorpc.Request) (*iorpc.Response, error) {
			request.Body.Close()
			return &iorpc.Response{
				Body: iorpc.Body{
					Size:   uint64(len(staticData)),
					Reader: staticData,
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

func (f *File) File() uintptr {
	return f.file.Fd()
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

func (b StaticBuffer) Buffer() [][]byte {
	return [][]byte{b}
}

func (b StaticBuffer) Close() error {
	return nil
}

func (b StaticBuffer) Read(p []byte) (n int, err error) {
	n = copy(p, b)
	return
}
