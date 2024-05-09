package spider

type Request struct {
	URL    string
	Method string
}

type Context struct {
	Body []byte
	Req  *Request
}
