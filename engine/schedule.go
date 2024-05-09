package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/Ysoding/pokemon-wiki-spider/spider"
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
	out chan spider.ParseResult
}

func NewEngine() *Crawler {
	return &Crawler{
		out: make(chan spider.ParseResult),
	}
}

func (c *Crawler) Run() {
}

func Run(ctx context.Context) error {
	fmt.Println("start")
	time.Sleep(time.Second * 2)
	return nil
}
