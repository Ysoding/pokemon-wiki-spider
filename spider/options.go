package spider

import "go.uber.org/zap"

type Option func(opts *Options)

type Options struct {
	Name     string
	URL      string
	Cookie   string
	WaitTime int64 // second
	MaxDepth int64
	logger   zap.Logger
	Fetcher  Fetcher
	Storage  Storage
}

var defaultOptions = Options{
	logger:   *zap.NewNop(),
	WaitTime: 5,
	MaxDepth: 5,
}

func WithName(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opts *Options) {
		opts.logger = *logger
	}
}

func WithURL(url string) Option {
	return func(opts *Options) {
		opts.URL = url
	}
}

func WithCookie(cookie string) Option {
	return func(opts *Options) {
		opts.Cookie = cookie
	}
}

func WithWaitTime(waitTime int64) Option {
	return func(opts *Options) {
		opts.WaitTime = waitTime
	}
}

func WithFetcher(fetcher Fetcher) Option {
	return func(opts *Options) {
		opts.Fetcher = fetcher
	}
}

func WithStorage(storage Storage) Option {
	return func(opts *Options) {
		opts.Storage = storage
	}
}
