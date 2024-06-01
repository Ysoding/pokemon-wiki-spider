package mongo

import (
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"go.uber.org/zap"
)

type Option func(opts *options)

type options struct {
	logger     *zap.Logger
	uri        string
	batchCount int
	dbName     string
}

var defaultOptions = options{
	logger:     zap.NewNop(),
	batchCount: global.DefaultBatchCount,
	dbName:     global.DefaultMongoDatabaseName,
}

func WithConnURI(url string) Option {
	return func(opts *options) { opts.uri = url }
}

func WithLogger(logger *zap.Logger) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func WithBatchCount(batchCount int) Option {
	return func(opts *options) {
		opts.batchCount = batchCount
	}
}

func WithDatabaseName(dbName string) Option {
	return func(opts *options) {
		opts.dbName = dbName
	}
}
