package main

import (
	"context"
	"sort"
	"time"
)

type IOStat struct {
	collectChan chan time.Duration
	statChan    chan chan *StatResult

	ctx    context.Context
	cancel func()
}

func NewIOStat(ctx context.Context, bufferSize int) *IOStat {
	ctx, cancel := context.WithCancel(ctx)
	ioStat := &IOStat{
		collectChan: make(chan time.Duration, bufferSize),
		statChan:    make(chan chan *StatResult, 1),
		ctx:         ctx,
		cancel:      cancel,
	}

	go func() {
		durations := make([]time.Duration, 0, bufferSize)
		for {
			select {
			case <-ctx.Done():
				return
			case dur := <-ioStat.collectChan:
				durations = append(durations, dur)
			case cb := <-ioStat.statChan:
				cb <- stat(durations)
				durations = durations[:0]
			}
		}
	}()

	return ioStat
}

func (ioStat *IOStat) Shutdown() {
	ioStat.cancel()
}

func (ioStat *IOStat) Collect(dur time.Duration) {
	ioStat.collectChan <- dur
}

func (ioStat *IOStat) Stat() *StatResult {
	cb := make(chan *StatResult, 1)
	ioStat.statChan <- cb
	return <-cb
}

type StatResult struct {
	Percent10, Percent25, Percent50, Percent75, Percent90, Percent95, Percent99 time.Duration
}

func stat(durations []time.Duration) *StatResult {
	if len(durations) == 0 {
		return &StatResult{}
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	result := &StatResult{}
	result.Percent10 = durations[len(durations)*1/10]
	result.Percent25 = durations[len(durations)*1/4]
	result.Percent50 = durations[len(durations)*1/2]
	result.Percent75 = durations[len(durations)*3/4]
	result.Percent90 = durations[len(durations)*9/10]
	result.Percent95 = durations[len(durations)*19/20]
	result.Percent99 = durations[len(durations)*99/100]
	return result
}

func (statResult *StatResult) String() string {
	return `
IO Latency distribution:
  10% in ` + statResult.Percent10.String() + `
  25% in ` + statResult.Percent25.String() + `
  50% in ` + statResult.Percent50.String() + `
  75% in ` + statResult.Percent75.String() + `
  90% in ` + statResult.Percent90.String() + `
  95% in ` + statResult.Percent95.String() + `
  99% in ` + statResult.Percent99.String() + `
`
}
