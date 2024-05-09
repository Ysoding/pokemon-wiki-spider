package spider

import "sync"

type TaskConfig struct {
	Name     string
	Cookie   string
	MaxDepth int64
}

type Task struct {
	Visited      map[string]bool
	VisistedLock sync.Mutex
	Rule         RuleTree
	Options
}
