package pokemon

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
	Index      int
	NameZh     string
	NameJa     string
	NameEn     string
	Form       string // 地区形态
	Type1      string
	Type2      string
	Generation int
}

var pokemonListTask = &spider.Task{
	Options: spider.Options{
		Name:     "pokemon_list",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					URL:      global.PokemonListURL,
					Method:   "GET",
					RuleName: "数据list",
				},
			}
			return roots, nil
		},

		Trunk: map[string]*spider.Rule{
			"数据list": {ParseFunc: ParsePokemonList},
		},
	},
}

func ParsePokemonList(ctx *spider.Context) (spider.ParseResult, error) {
	var items []*Data

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	locations := []string{
		"关都",
		"城都",
		"丰缘",
		"神奥",
		"合众",
		"卡洛斯",
		"阿罗拉",
		"伽勒尔",
		"帕底亚",
	}

	for idx, location := range locations {
		items = append(items, getData(doc, location, idx+1)...)
	}

	var result []interface{}
	for _, value := range items {
		result = append(result, value)
	}

	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     result,
	}, nil
}

func parseElement(ele *goquery.Selection, generation int) (*Data, error) {

	indexStr := strings.TrimSpace(ele.Find("td").Eq(0).Text())
	index, err := strconv.Atoi(strings.ReplaceAll(indexStr, "#", ""))
	if err != nil {
		return nil, fmt.Errorf("parse index err: %v", err)
	}

	nameZh := strings.TrimSpace(ele.Find("td").Eq(3).Find("a").Text())
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

	data := &Data{
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

func getData(doc *goquery.Document, locationName string, generation int) []*Data {
	var res []*Data

	doc.Find(".s-" + locationName + " > tbody > tr").Each(func(i int, s *goquery.Selection) {
		if i < 2 {
			return
		}

		data, err := parseElement(s, generation)
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, data)
	})

	return res
}
