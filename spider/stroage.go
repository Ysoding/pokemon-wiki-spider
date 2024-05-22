package spider

type Storage interface {
	Save(data interface{}) error
}
