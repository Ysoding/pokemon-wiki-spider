package ability

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

type PokemonAbilityData struct {
	Index       int
	NameZh      string
	Form        string // 地区形态
	Type1       string
	Type2       string
	Ability1    string
	Ability2    string
	HideAbility string
	Generation  int
}

var PokemonAbilityListTask = &spider.Task{
	Options: spider.Options{
		Name:     "pokemon_ability_list",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					URL:      global.PokemonAbilityListURL,
					Method:   "GET",
					RuleName: "list",
				},
			}
			return roots, nil
		},

		Trunk: map[string]*spider.Rule{
			"list": {ParseFunc: parsePokemonAbilityList},
		},
	},
}

func parsePokemonAbilityList(ctx *spider.Context) (spider.ParseResult, error) {
	var items []*PokemonAbilityData

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	for idx, location := range global.LocationNameList {
		items = append(items, getPokemonAbilityData(doc, location, idx+1)...)
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

func parsePokemonAbilityElement(ele *goquery.Selection, generation int) (*PokemonAbilityData, error) {

	indexStr := strings.TrimSpace(ele.Children().Eq(0).Text())
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("parse index err: %v", err)
	}

	nameZh := strings.TrimSpace(ele.Children().Eq(2).Text())
	smallEle := ele.Children().Eq(2).Find("small")
	form := ""
	if smallEle != nil {
		form = strings.TrimSpace(smallEle.Text())
	}

	type1 := strings.TrimSpace(ele.Children().Eq(3).Text())
	type2Ele := ele.Children().Eq(4)
	type2 := ""
	if !type2Ele.HasClass("hide") {
		type2 = strings.TrimSpace(type2Ele.Text())
	}

	ability1 := strings.TrimSpace(ele.Children().Eq(5).Text())
	ability2Ele := ele.Children().Eq(6)
	ability2 := ""
	if !ability2Ele.HasClass("hide") {
		ability2 = strings.TrimSpace(ability2Ele.Text())
	}

	hideAbilityEle := ele.Children().Eq(7).Find("a")
	hideAbility := "无"
	if hideAbilityEle != nil {
		ability2 = strings.TrimSpace(hideAbilityEle.Text())
	}

	data := &PokemonAbilityData{
		Index:       index,
		NameZh:      nameZh,
		Form:        form,
		Type1:       type1,
		Type2:       type2,
		Ability1:    ability1,
		Ability2:    ability2,
		HideAbility: hideAbility,
		Generation:  generation,
	}

	return data, nil
}

func getPokemonAbilityData(doc *goquery.Document, locationName string, generation int) []*PokemonAbilityData {
	var res []*PokemonAbilityData

	doc.Find(".bg-" + locationName + " > tbody > tr").Each(func(i int, s *goquery.Selection) {
		if i < 2 {
			return
		}

		data, err := parsePokemonAbilityElement(s, generation)
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, data)
	})

	return res
}
