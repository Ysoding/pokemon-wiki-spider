package mongo

import (
	"github.com/Ysoding/pokemon-wiki-spider/db/mongodb"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

type MongoStorage struct {
	dataDocker []*spider.DataCell // cache
	db         mongodb.DBer
	options
}

func (m *MongoStorage) Save(datas ...*spider.DataCell) error {
	for _, data := range datas {
		if len(m.dataDocker) >= m.batchCount {
			if err := m.Flush(); err != nil {
				return err
			}
		}

		m.dataDocker = append(m.dataDocker, data)
	}
	return nil
}

func (m *MongoStorage) Flush() error {
	if len(m.dataDocker) == 0 {
		return nil
	}
	m.logger.Info("mongo storage start flush data")

	defer func() {
		m.dataDocker = nil
	}()

	data := make([]interface{}, 0)

	for _, d := range m.dataDocker {
		nd := d.Data["Data"].(map[string]interface{})
		data = append(data, nd)
	}

	return m.db.InsertMany(mongodb.TableData{
		TableName: m.dataDocker[0].GetTableName(),
		Data:      data,
	})
}

func New(opts ...Option) (*MongoStorage, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	s := &MongoStorage{options: options}

	var err error
	s.db, err = mongodb.New(mongodb.WithConnURI(s.uri),
		mongodb.WithLogger(s.logger),
		mongodb.WithDatabaseName(s.dbName))
	if err != nil {
		return nil, err
	}

	return s, nil
}
