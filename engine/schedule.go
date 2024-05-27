package engine

import (
	"runtime/debug"
	"sync"

	"github.com/Ysoding/pokemon-wiki-spider/spider"
	"go.uber.org/zap"
)

type Scheduler interface {
	Schedule()
	Push(...*spider.Request)
	Pull() *spider.Request
	Close()
}

type Crawler struct {
	out          chan spider.ParseResult
	visisted     map[string]bool
	visistedLock sync.Mutex

	failures     map[string]*spider.Request // id -> request
	failuresLock sync.Mutex

	options
}

func NewEngine(opts ...Option) *Crawler {

	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	c := &Crawler{
		out:      make(chan spider.ParseResult),
		visisted: make(map[string]bool),
		failures: make(map[string]*spider.Request),
		options:  options,
	}

	return c
}

func (c *Crawler) handleSeeds() {
	res := []*spider.Request{}
	for _, task := range c.Seeds {
		c.Logger.Info("parse task", zap.String("Name", task.Name))

		task.Fetcher = c.Fetcher

		reqs, err := task.Rule.Root()
		if err != nil {
			c.Logger.Error("got task root failed", zap.Error(err))
			continue
		}

		for _, req := range reqs {
			c.Logger.Info("request", zap.String("URL", req.URL))
			req.Task = task
		}

		res = append(res, reqs...)
	}
	go c.scheduler.Push(res...)
	c.Logger.Info("parse task done")
}

func (c *Crawler) Run() error {
	go c.schedule()
	var wg sync.WaitGroup

	for i := 0; i < c.WorkerCount; i++ {
		wg.Add(1)
		go c.createWorker(&wg)
	}

	go c.handleSeeds()
	go c.handleResult()
	wg.Wait()
	return nil
}

func (c *Crawler) Shutdown() {
	c.scheduler.Close()
}

func (c *Crawler) handleResult() {
	for result := range c.out {
		for _, item := range result.Items {
			c.Logger.Sugar().Infow("crawler", "got item", item)
		}
	}
}

func (c *Crawler) hashVisited(req *spider.Request) bool {
	c.visistedLock.Lock()
	defer c.visistedLock.Unlock()
	return c.visisted[req.Unique()]
}

func (c *Crawler) storeVisited(reqs ...*spider.Request) {
	c.visistedLock.Lock()
	defer c.visistedLock.Unlock()
	for _, r := range reqs {
		c.visisted[r.Unique()] = true
	}
}

func (c *Crawler) setFailure(req *spider.Request) {
	c.failuresLock.Lock()
	defer c.failuresLock.Unlock()
	if _, ok := c.failures[req.Unique()]; !ok {
		c.failures[req.Unique()] = req
		c.scheduler.Push(req)
	}
}

func (c *Crawler) createWorker(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			c.Logger.Sugar().Errorw("worker", "err", err, "stack", string(debug.Stack()))
		}
	}()
	defer wg.Done()

	for {
		req := c.scheduler.Pull()
		if req == nil {
			return
		}

		c.Logger.Info("start parse req", zap.String("URL", req.URL))
		if err := req.Check(); err != nil {
			c.Logger.Debug("request check failed", zap.Error(err))
		}

		if c.hashVisited(req) {
			c.Logger.Debug("requst has visisted ", zap.String("url", req.URL))
			continue
		}

		c.storeVisited(req)

		c.Logger.Info("start fetch body", zap.String("URL", req.URL))
		body, err := req.Fetch()
		if err != nil {
			c.Logger.Error("can't fetch ",
				zap.Error(err),
				zap.String("url", req.URL),
			)
			c.setFailure(req)
			continue
		}

		if len(body) < 6000 {
			c.Logger.Error("can't fetch not correct length ",
				zap.Int("length", len(body)),
				zap.String("url", req.URL))
			c.setFailure(req)
			continue
		}

		c.Logger.Info("start call parse func", zap.String("URL", req.URL))
		rule := req.Task.Rule.Trunk[req.RuleName]
		result, err := rule.ParseFunc(&spider.Context{
			Body: body,
			Req:  req,
		})

		if err != nil {
			c.Logger.Error("ParseFunc failed ", zap.String("url", req.URL), zap.Error(err))
			continue
		}

		if len(result.Requesrts) > 0 {
			go c.scheduler.Push(result.Requesrts...)
		}
		c.out <- result

		c.Logger.Info("parse req done", zap.String("URL", req.URL))
	}
}

func (c *Crawler) schedule() {
	c.scheduler.Schedule()
}

type Schedule struct {
	requestCh chan *spider.Request
	workerCh  chan *spider.Request
	reqQueue  []*spider.Request
}

func NewSchedule() *Schedule {
	s := &Schedule{
		requestCh: make(chan *spider.Request),
		workerCh:  make(chan *spider.Request),
		reqQueue:  make([]*spider.Request, 0),
	}

	return s
}

func (s *Schedule) Schedule() {
	var ch chan *spider.Request
	var req *spider.Request
	for {
		if req == nil && len(s.reqQueue) > 0 {
			req = s.reqQueue[0]
			s.reqQueue = s.reqQueue[1:]
			ch = s.workerCh
		}

		select {
		case r := <-s.requestCh:
			if r == nil {
				return
			}
			// TODO: 这里可以做优先级队列
			s.reqQueue = append(s.reqQueue, r)
		case ch <- req:
			req = nil
			ch = nil
		}
	}
}

func (s *Schedule) Close() {
	close(s.requestCh)
	close(s.workerCh)
}

func (s *Schedule) Push(requests ...*spider.Request) {
	for _, req := range requests {
		s.requestCh <- req
	}
}

func (s *Schedule) Pull() *spider.Request {
	r := <-s.workerCh
	return r
}
