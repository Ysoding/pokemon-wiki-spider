package engine

import (
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
	"go.uber.org/zap"
)

type Option func(opts *options)

type options struct {
	WorkerCount int
	Fetcher     spider.Fetcher
	Storage     spider.Storage
	Seeds       []*spider.Task
	scheduler   Scheduler
	Logger      *zap.Logger
}

var defaultOptions = options{
	WorkerCount: global.DefaultWorkerCount,
	Logger:      zap.NewNop(),
}

func WithLogger(l *zap.Logger) Option {
	return func(opts *options) {
		opts.Logger = l
	}
}

func WithStorage(s spider.Storage) Option {
	return func(opts *options) {
		opts.Storage = s
	}
}

func WithWorkerCount(n int) Option {
	return func(opts *options) {
		opts.WorkerCount = n
	}
}

func WithSeeds(seeds []*spider.Task) Option {
	return func(opts *options) {
		opts.Seeds = seeds
	}
}

func WithScheduler(scheduler Scheduler) Option {
	return func(opts *options) {
		opts.scheduler = scheduler
	}
}

func WithFetcher(fetcher spider.Fetcher) Option {
	return func(opts *options) {
		opts.Fetcher = fetcher
	}
}
