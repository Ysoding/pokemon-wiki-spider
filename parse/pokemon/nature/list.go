package nature

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

type Data struct {
	NameZh            string
	NameJa            string
	NameEn            string
	EasyGrowthAbility string
	HardGrowthAbility string
	FavoriteTaste     string
	DislikedTaste     string
}

var PokemonNatureListTask = &spider.Task{
	Options: spider.Options{
		Name:     "nature_list",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					URL:      global.PokemonNatureListURL,
					Method:   "GET",
					RuleName: "list",
				},
			}
			return roots, nil
		},

		Trunk: map[string]*spider.Rule{
			"list": {ParseFunc: ParsePokemonNatureList},
		},
	},
}

func ParsePokemonNatureList(ctx *spider.Context) (spider.ParseResult, error) {
	var items []*Data

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	items = append(items, getData(doc)...)

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
	nameZh := strings.TrimSpace(ele.Children().Eq(0).Text())
	nameJa := strings.TrimSpace(ele.Children().Eq(1).Text())
	nameEn := strings.TrimSpace(ele.Children().Eq(2).Text())

	easyGrowthAbility := strings.TrimSpace(ele.Children().Eq(3).Text())
	hardGrowthAbility := strings.TrimSpace(ele.Children().Eq(4).Text())
	favoriteTaste := strings.TrimSpace(ele.Children().Eq(5).Text())
	dislikedTaste := strings.TrimSpace(ele.Children().Eq(6).Text())

	data := &Data{
		NameZh:            nameZh,
		NameJa:            nameJa,
		NameEn:            nameEn,
		EasyGrowthAbility: easyGrowthAbility,
		HardGrowthAbility: hardGrowthAbility,
		FavoriteTaste:     favoriteTaste,
		DislikedTaste:     dislikedTaste,
	}

	return data, nil
}

func getData(doc *goquery.Document) []*Data {
	var res []*Data

	doc.Find("#mw-content-text table").Eq(0).Find("tbody > tr").Each(func(i int, s *goquery.Selection) {
		data, err := parseElement(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, data)
	})

	return res
}
