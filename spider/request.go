package spider

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"
)

type Request struct {
	Task     *Task
	URL      string
	Method   string
	RuleName string
	Depth    int64
}

type Context struct {
	Body []byte
	Req  *Request
}

func (r *Request) Fetch() ([]byte, error) {
	sleepTime := rand.Int63n(r.Task.WaitTime * 1000)
	time.Sleep(time.Duration(sleepTime))

	return r.Task.Fetcher.Get(r)
}

func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.URL + r.Method))
	return hex.EncodeToString(block[:])
}

func (r *Request) Check() error {
	if r.Depth > r.Task.MaxDepth {
		return errors.New("max depth limit reached")
	}
	return nil
}
