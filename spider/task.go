package spider

import "sync"

type Task struct {
	Visited      map[string]bool
	VisistedLock sync.Mutex
	Rule         RuleTree
	Options
}

type Fetcher interface {
	Get(req *Request) ([]byte, error)
}

func NewTask(opts ...Option) *Task {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	t := &Task{Options: options}

	return t
}
