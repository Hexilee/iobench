package gorpcbench

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/valyala/gorpc"
)

var filePool = sync.Pool{
	New: func() any {
		file, err := os.OpenFile("../data/data", os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		return &File{file}
	},
}

type File struct {
	*os.File
}

func (f *File) Close() error {
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}
	filePool.Put(f)
	return nil
}

func (f *File) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, f.File)
}

func ListenAndServe(addr string) error {
	// Start rpc server serving registered service.
	s := &gorpc.Server{
		// Accept clients on this TCP address.
		Addr: addr,

		// Echo handler - just return back the message we received from the client
		Handler: func(clientAddr string, request gorpc.Request) gorpc.Response {
			if request.Body != nil {
				request.Body.Close()
			}
			switch ret := filePool.Get().(type) {
			case error:
				buf := bytes.NewBufferString(ret.Error())
				return gorpc.Response{
					Size: uint64(buf.Len()),
					Body: io.NopCloser(buf),
				}
			case *File:
				stat, err := ret.Stat()
				if err != nil {
					buf := bytes.NewBufferString(err.Error())
					return gorpc.Response{
						Size: uint64(buf.Len()),
						Body: io.NopCloser(buf),
					}
				}
				return gorpc.Response{
					Size: uint64(stat.Size()),
					Body: ret,
				}
			default:
				panic("unreachable")
			}
		},
	}
	s.CloseBody = true
	return s.Serve()
}
