package move

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

type Data struct {
	Index       int
	NameZh      string
	NameJa      string
	NameEn      string
	Type        string
	Category    string
	Power       string
	Accuracy    string
	PP          string
	Description string
	Generation  int
}

var MoveListTask = &spider.Task{
	Options: spider.Options{
		Name:     "move_list",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					URL:      global.PokemonMoveListURL,
					Method:   "GET",
					RuleName: "list",
				},
			}
			return roots, nil
		},

		Trunk: map[string]*spider.Rule{
			"list": {ParseFunc: ParsePokemonMoveList},
		},
	},
}

func ParsePokemonMoveList(ctx *spider.Context) (spider.ParseResult, error) {
	var items []*Data

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	for idx, location := range global.LocationNameList {
		items = append(items, getData(doc, location, idx+1)...)
	}

	var result []interface{}
	for _, value := range items {
		result = append(result, ctx.Output(global.StructToMap(value)))
	}

	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     result,
	}, nil
}

func parseElement(ele *goquery.Selection) (*Data, error) {

	indexStr := strings.TrimSpace(ele.Children().Eq(0).Text())
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("parse index err: %v", err)
	}

	nameZh := strings.TrimSpace(ele.Children().Eq(1).Text())
	nameJa := strings.TrimSpace(ele.Children().Eq(2).Text())
	nameEn := strings.TrimSpace(ele.Children().Eq(3).Text())
	typ := strings.TrimSpace(ele.Children().Eq(4).Text())
	category := strings.TrimSpace(ele.Children().Eq(5).Text())
	power := strings.TrimSpace(ele.Children().Eq(6).Text())
	accuracy := strings.TrimSpace(ele.Children().Eq(7).Text())
	pp := strings.TrimSpace(ele.Children().Eq(8).Text())
	desc := strings.TrimSpace(ele.Children().Eq(9).Text())

	data := &Data{
		Index:       index,
		NameZh:      nameZh,
		NameJa:      nameJa,
		NameEn:      nameEn,
		Type:        typ,
		Category:    category,
		Power:       power,
		Accuracy:    accuracy,
		PP:          pp,
		Description: desc,
	}

	return data, nil
}

func getData(doc *goquery.Document, locationName string, generation int) []*Data {
	var res []*Data

	doc.Find(".bg-" + locationName + " > tbody > tr").Each(func(i int, s *goquery.Selection) {
		data, err := parseElement(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		data.Generation = generation
		res = append(res, data)
	})

	return res
}
