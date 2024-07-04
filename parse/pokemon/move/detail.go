package move

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Ysoding/pokemon-wiki-spider/db/mongodb"
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/limiter"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type MoveDetailData struct {
	Index  int
	NameZh string
	Desc   string
	ImgUrl string
	Notes  string
	Scope  string
	Effect string
}

var MoveDetailTask = &spider.Task{
	Options: spider.Options{
		Name:     "move_detail",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
		Limit: limiter.Multi(
			rate.NewLimiter(limiter.Per(1, 1*time.Second), 1),
		),
	},
	Rule: spider.RuleTree{
		Root: roots,
		Trunk: map[string]*spider.Rule{
			"parse": {ParseFunc: parseMoveDetail},
		},
	},
}

func parseMoveDetail(ctx *spider.Context) (spider.ParseResult, error) {
	index, _ := ctx.Req.TempData.Get("index").(int)
	nameZh, _ := ctx.Req.TempData.Get("nameZh").(string)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	trList := doc.Find("#mw-content-text > .mw-parser-output > .roundy").First().Find("tbody").Children()
	desc := trList.Eq(1).Text()

	imgUrl := ""
	if trList.Eq(2).Find("span").First().Length() != 0 {
		imgUrl = trList.Eq(2).Find("span").AttrOr("data-url", "")
		imgUrl = strings.Replace(imgUrl, "//media.52poke.com", "https://s1.52poke.wiki", 1)
	} else if trList.Eq(2).Find("img").First().Length() != 0 {
		imgUrl = trList.Eq(2).Find("img").AttrOr("data-url", "")
		imgUrl = strings.Replace(imgUrl, "//media.52poke.com", "https://s1.52poke.wiki", 1)
		decodedUrl, _ := url.QueryUnescape(imgUrl)
		imgUrl = decodedUrl
	}

	tmp := trList.Eq(3).Find("table > tbody > tr")
	notes := tmp.Eq(7).Find("td > div > ul").Text()
	scope := tmp.Eq(9).Text()

	// 招式附加效果
	ele := doc.Find("#mw-content-text > .mw-parser-output > h2").Eq(0).Next()
	effect := ""
	for ele.Length() > 0 && ele.Nodes[0].Data != "h2" {
		effect += ele.Text()
		ele = ele.Next()
	}

	var result []interface{}
	result = append(result, ctx.Output(global.StructToMap(&MoveDetailData{
		Index:  index,
		NameZh: nameZh,
		Desc:   desc,
		ImgUrl: imgUrl,
		Notes:  notes,
		Scope:  scope,
		Effect: effect,
	})))

	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     result,
	}, nil
}

func roots() ([]*spider.Request, error) {
	db, err := mongodb.New(mongodb.WithConnURI(os.Getenv("MONGO_URL")),
		mongodb.WithDatabaseName(global.DefaultMongoDatabaseName))
	if err != nil {
		return nil, err
	}

	var moveListData []MoveListData
	if err := db.Find(global.PokemonMoveListName, bson.D{}, &moveListData); err != nil {
		return nil, err
	}

	var requesrts []*spider.Request
	for _, d := range moveListData {
		req := &spider.Request{
			URL:      fmt.Sprintf("https://wiki.52poke.com/zh-hans/%s（招式）", d.NameZh),
			Method:   "GET",
			RuleName: "parse",
		}
		req.TempData = &spider.TempData{}

		if err := req.TempData.Set("index", d.Index); err != nil {
			zap.L().Error("set temp data error", zap.Error(err))
			continue
		}

		if err := req.TempData.Set("nameZh", d.NameZh); err != nil {
			zap.L().Error("set temp data error", zap.Error(err))
			continue
		}

		requesrts = append(requesrts, req)
	}

	return requesrts, nil
}
