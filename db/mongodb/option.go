package mongodb

import "go.uber.org/zap"

type Option func(opts *options)

type options struct {
	logger *zap.Logger
	uri    string
	dbName string
}

var defaultOptions = options{
	logger: zap.NewNop(),
}

func WithConnURI(url string) Option {
	return func(opts *options) { opts.uri = url }
}

func WithLogger(logger *zap.Logger) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func WithDatabaseName(dbName string) Option {
	return func(opts *options) {
		opts.dbName = dbName
	}
}
