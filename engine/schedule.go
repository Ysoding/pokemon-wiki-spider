package engine

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
	"go.uber.org/zap"
)

type Scheduler interface {
	Schedule()
	Push(...*spider.Request)
	Pull() *spider.Request
}

type CrawlerStore struct {
	list []*spider.Task
}

type Crawler struct {
	out          chan spider.ParseResult
	Visisted     map[string]bool
	VisistedLock sync.Mutex

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
		Visisted: make(map[string]bool),
		failures: make(map[string]*spider.Request),
		options:  options,
	}

	return c
}

func Run(ctx context.Context, logger *zap.Logger) error {
	seeds := []*spider.Task{
		spider.NewTask(spider.WithLogger(logger),
			spider.WithURL(global.PokemonListURL)),
	}
	e := NewEngine(WithLogger(logger), WithScheduler(NewSchedule()), WithSeeds(seeds))
	return e.Run()
}

func (c *Crawler) handleSeeds() {
	// TODO
	for _, task := range c.Seeds {
		reqs, err := task.Rule.Root()
	}
}

func (c *Crawler) Run() error {
	go c.handleSeeds()
	go c.Schedule()

	for i := 0; i < c.WorkerCount; i++ {
		go c.CreateWorker()
	}

	go c.HandleResult()
	return nil
}

func (c *Crawler) HandleResult() {
	for result := range c.out {
		for _, item := range result.Items {
			c.Logger.Sugar().Infow("crawler", "got item", item)
		}
	}
}

func (c *Crawler) CreateWorker() {
	defer func() {
		if err := recover(); err != nil {
			c.Logger.Sugar().Errorw("worker", "err", err, "stack", string(debug.Stack()))
		}
	}()

	for {
		req := c.scheduler.Pull()
		c.Logger.Sugar().Infow("worker", "req", req)
	}
}

func (c *Crawler) Schedule() {
	c.scheduler.Schedule()
}

type Schedule struct {
	requestCh chan *spider.Request
	workerCh  chan *spider.Request
	reqQueue  []*spider.Request
	Logger    *zap.Logger
}

func NewSchedule() *Schedule {
	s := &Schedule{}

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
			// TODO: 这里可以做优先级
			s.reqQueue = append(s.reqQueue, r)
		case ch <- req:
			req = nil
			ch = nil
		}
	}
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
