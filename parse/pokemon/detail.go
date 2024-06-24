package pokemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Ysoding/pokemon-wiki-spider/db/mongodb"
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
	"go.mongodb.org/mongo-driver/bson"
)

type PokemonDetailData struct {
	Index      int
	NameZh     string
	NameJa     string
	NameEn     string
	Form       string // 地区形态
	Type1      string
	Type2      string
	Generation int
}

var PokemonDetailTask = &spider.Task{
	Options: spider.Options{
		Name:     global.PokemonDetailTaskName,
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			return roots()
		},

		Trunk: map[string]*spider.Rule{
			"parse": {ParseFunc: parsePokemonDetail},
		},
	},
}

func parsePokemonDetail(ctx *spider.Context) (spider.ParseResult, error) {
	// TODO: parse
	// var items []*PokemonDetailData

	// doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	// if err != nil {
	// 	return spider.ParseResult{}, err
	// }

	// for idx, location := range global.LocationNameList {
	// 	items = append(items, getPokemonDetailData(doc, location, idx+1)...)
	// }

	// var result []interface{}
	// for _, value := range items {
	// 	result = append(result, ctx.Output(global.StructToMap(value)))
	// }

	fmt.Println("TODO")
	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     make([]interface{}, 0),
	}, nil
}

func roots() ([]*spider.Request, error) {
	// select all pokemon from mongodb
	// create new parse detail task for every pokemon
	db, err := mongodb.New(mongodb.WithConnURI(os.Getenv("MONGO_URL")),
		mongodb.WithDatabaseName(global.DefaultMongoDatabaseName))
	if err != nil {
		return nil, err
	}

	var pokmonListData []PokemonListData
	if err := db.Find(global.PokemonListTaskName, bson.D{}, &pokmonListData); err != nil {
		return nil, err
	}

	var requesrts []*spider.Request
	for _, d := range pokmonListData {
		requesrts = append(requesrts, &spider.Request{
			URL:      fmt.Sprintf("https://wiki.52poke.com/zh-hans/%s", d.NameZh),
			Method:   "GET",
			RuleName: "parse",
		})
	}

	return requesrts, nil
}

func parsePokemonDetailElement(ele *goquery.Selection, generation int) (*PokemonDetailData, error) {

	indexStr := strings.TrimSpace(ele.Find("td").Eq(0).Text())
	index, err := strconv.Atoi(strings.ReplaceAll(indexStr, "#", ""))
	if err != nil {
		return nil, fmt.Errorf("parse index err: %v", err)
	}

	nameZh := strings.TrimSpace(ele.Find("td").Eq(3).Text())
	smallEle := ele.Find("td").Eq(3).Find("small")
	form := ""
	if smallEle != nil {
		form = strings.TrimSpace(smallEle.Text())
	}

	nameJa := strings.TrimSpace(ele.Find("td").Eq(4).Text())
	nameEn := strings.TrimSpace(ele.Find("td").Eq(5).Text())

	type1 := strings.TrimSpace(ele.Find("td").Eq(6).Text())
	type2Ele := ele.Find("td").Eq(7)
	type2 := ""
	if !type2Ele.HasClass("hide") {
		type2 = strings.TrimSpace(type2Ele.Text())
	}

	data := &PokemonDetailData{
		Index:      index,
		NameZh:     nameZh,
		NameJa:     nameJa,
		NameEn:     nameEn,
		Form:       form,
		Type1:      type1,
		Type2:      type2,
		Generation: generation,
	}

	return data, nil
}

func getPokemonDetailData(doc *goquery.Document, locationName string, generation int) []*PokemonDetailData {
	var res []*PokemonDetailData

	doc.Find(".s-" + locationName + " > tbody > tr").Each(func(i int, s *goquery.Selection) {
		if i < 2 {
			return
		}

		data, err := parsePokemonDetailElement(s, generation)
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, data)
	})

	return res
}
