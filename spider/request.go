package spider

import "math/rand"

type Request struct {
	Task *Task
	URL    string
	Method string
}

type Context struct {
	Body []byte
	Req  *Request
}


func (r *Request) Fetch() ([]byte, error) {
	sleepTime := rand.Int63n(r.)
}