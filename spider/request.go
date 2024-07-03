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
	TempData *TempData
}

type Context struct {
	Body []byte
	Req  *Request
}

func (c *Context) Output(data interface{}) *DataCell {
	res := &DataCell{
		Task: c.Req.Task,
	}

	res.Data = make(map[string]interface{})
	res.Data["Task"] = c.Req.Task.Name
	res.Data["Data"] = data
	res.Data["Time"] = time.Now().Format("2006-01-02 15:04:05")

	return res
}

func (r *Request) Fetch() ([]byte, error) {
	sleepTime := rand.Int63n(r.Task.WaitTime * 1000)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)

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
