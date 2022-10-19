module github.com/hexilee/iobench/go

go 1.19

replace github.com/hexilee/iorpc => ../../iorpc

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/felixge/fgprof v0.9.3
	github.com/google/uuid v1.3.0
	github.com/hanwen/go-fuse v1.0.0
	github.com/hexilee/iorpc v0.0.0-20221008174116-71f3f8f795c5
	github.com/valyala/fasthttp v1.40.0
	golang.org/x/net v0.0.0-20221004154528-8021a29435af
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/text v0.3.7 // indirect
)
