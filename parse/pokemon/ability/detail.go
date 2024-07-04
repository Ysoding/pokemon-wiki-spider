package ability

import (
	"bytes"
	"fmt"
	"os"
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

type AbilityDetailData struct {
	Index  int
	NameZh string
	Desc   string
	Effect string
	Owners []string // 拥有此特性的宝可梦
}

var MoveDetailTask = &spider.Task{
	Options: spider.Options{
		Name:     "ability_detail",
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
			"parse": {ParseFunc: parseAbilityDetail},
		},
	},
}

func parseAbilityDetail(ctx *spider.Context) (spider.ParseResult, error) {
	index, _ := ctx.Req.TempData.Get("index").(int)
	nameZh, _ := ctx.Req.TempData.Get("nameZh").(string)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	desc := doc.Find("#mw-content-text > .mw-parser-output > .at-c > tbody").Eq(0).Find("tr").Eq(4).Text()

	ele := doc.Find("#mw-content-text > .mw-parser-output > h2").Eq(0).Next()
	effect := ""
	for ele.Length() > 0 && ele.Nodes[0].Data != "h2" {
		effect += ele.Text()
		ele = ele.Next()
	}

	var pokemonList []string
	doc.Find("#具有该特性的宝可梦").Parent().Next().Find("tr").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("bgwhite") {
			pokemon := s.Children().Eq(2).Children().Eq(0).Text()
			if !contains(pokemonList, pokemon) {
				pokemonList = append(pokemonList, pokemon)
			}
		}
	})

	var result []interface{}
	result = append(result, ctx.Output(global.StructToMap(&AbilityDetailData{
		Index:  index,
		NameZh: nameZh,
		Desc:   desc,
		Effect: effect,
		Owners: pokemonList,
	})))

	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     result,
	}, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func roots() ([]*spider.Request, error) {
	db, err := mongodb.New(mongodb.WithConnURI(os.Getenv("MONGO_URL")),
		mongodb.WithDatabaseName(global.DefaultMongoDatabaseName))
	if err != nil {
		return nil, err
	}

	var abilityListData []AbilityData
	if err := db.Find(global.PokemonAbilityListName, bson.D{}, &abilityListData); err != nil {
		return nil, err
	}

	var requesrts []*spider.Request
	for _, d := range abilityListData {
		req := &spider.Request{
			URL:      fmt.Sprintf("https://wiki.52poke.com/zh-hans/%s（特性）", d.NameZh),
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
