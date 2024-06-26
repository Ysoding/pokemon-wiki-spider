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

type AbilityData struct {
	Index       int
	NameZh      string
	NameJa      string
	NameEn      string
	Description string
	CommonCnt   int
	HiddenCnt   int
	Generation  int
}

var AbilityListTask = &spider.Task{
	Options: spider.Options{
		Name:     "ability_list",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 0,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					URL:      global.AbilityListURL,
					Method:   "GET",
					RuleName: "list",
				},
			}
			return roots, nil
		},

		Trunk: map[string]*spider.Rule{
			"list": {ParseFunc: parseAbilityList},
		},
	},
}

func parseAbilityList(ctx *spider.Context) (spider.ParseResult, error) {
	var items []*AbilityData

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	for idx, location := range global.LocationNameList {
		if idx < 2 {
			continue
		}
		if location == "阿罗拉" {
			location = "阿羅拉"
		}
		items = append(items, getAbilityData(doc, location, idx+1)...)
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

func parseAbilityElement(ele *goquery.Selection, generation int) (*AbilityData, error) {

	indexStr := strings.TrimSpace(ele.Children().Eq(0).Text())
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("parse index err: %v", err)
	}

	nameZh := strings.TrimSpace(ele.Children().Eq(1).Text())
	nameJa := strings.TrimSpace(ele.Children().Eq(2).Text())
	nameEn := strings.TrimSpace(ele.Children().Eq(3).Text())

	description := strings.TrimSpace(ele.Children().Eq(4).Text())

	commonCntStr := strings.TrimSpace(ele.Children().Eq(5).Text())
	commonCnt, err := strconv.Atoi(commonCntStr)
	if err != nil {
		return nil, fmt.Errorf("parse 常见 fail: %v", err)
	}

	hiddenCntStr := strings.TrimSpace(ele.Children().Eq(6).Text())
	hiddenCnt, err := strconv.Atoi(hiddenCntStr)
	if err != nil {
		return nil, fmt.Errorf("parse 隐藏 fail: %v", err)
	}

	data := &AbilityData{
		Index:       index,
		NameZh:      nameZh,
		NameJa:      nameJa,
		NameEn:      nameEn,
		Description: description,
		CommonCnt:   commonCnt,
		HiddenCnt:   hiddenCnt,
		Generation:  generation,
	}

	return data, nil
}

func getAbilityData(doc *goquery.Document, locationName string, generation int) []*AbilityData {
	var res []*AbilityData

	selector := ".s-" + locationName + " > tbody > tr"
	if locationName == "帕底亚" {
		selector = ".b-" + locationName + " > tbody > tr"
	}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if i < 2 {
			return
		}

		data, err := parseAbilityElement(s, generation)
		if err != nil {
			fmt.Println(err)
			return
		}
		res = append(res, data)
	})

	return res
}
