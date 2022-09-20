package main

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var (
	fastSlowStat = NewIOStat(context.TODO(), 10000)
	fastFastStat = NewIOStat(context.TODO(), 100000)
	fastMockStat = NewIOStat(context.TODO(), 100000)
)

func fastHttpHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Request.URI().Path()) {
	case "/fast":
		fastFastHandler(ctx)
	case "/slow":
		fastSlowHandler(ctx)
	case "/mock":
		fastMockHandler(ctx)
	case "/stat/fast":
		fastStatHandler(fastFastStat)(ctx)
	case "/stat/slow":
		fastStatHandler(fastSlowStat)(ctx)
	case "/stat/mock":
		fastStatHandler(fastMockStat)(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func fastMockHandler(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	data := mockRead()
	fastMockStat.Collect(time.Since(start))
	ctx.Write(data)
}

func fastFastHandler(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	data, err := os.ReadFile("../data/data")
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	fastFastStat.Collect(time.Since(start))
	ctx.Write(data)
}

func fastSlowHandler(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	data, err := func() ([]byte, error) {
		filepath := path.Join("../data/tmp/", uuid.New().String())
		file, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		data, err := os.ReadFile("../data/data")
		if err != nil {
			return nil, err
		}

		_, err = file.Write(data)
		if err != nil {
			return nil, err
		}
		err = file.Sync()
		if err != nil {
			return nil, err
		}
		return data, os.Remove(filepath)
	}()

	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	fastSlowStat.Collect(time.Since(start))
	ctx.Write(data)
}

func fastStatHandler(ioStat *IOStat) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		stat := ioStat.Stat()
		ctx.WriteString(stat.String())
	}
}
