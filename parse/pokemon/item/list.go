package item

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
	NameZh      string
	NameJa      string
	NameEn      string
	Type        string
	Description string
	ImageURL    string
}

var ItemListTask = &spider.Task{
	Options: spider.Options{
		Name:     "item_list",
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					URL:      global.PokemonItemListURL,
					Method:   "GET",
					RuleName: "list",
				},
			}
			return roots, nil
		},

		Trunk: map[string]*spider.Rule{
			"list": {ParseFunc: ParsePokemonItemList},
		},
	},
}

func ParsePokemonItemList(ctx *spider.Context) (spider.ParseResult, error) {
	var items []*Data

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	items = append(items, getData(doc, "野外使用的道具", "道具#野外使用的道具")...)
	items = append(items, getData(doc, "培养宝可梦的道具", "道具#培养宝可梦的道具")...)
	items = append(items, getData(doc, "进化道具", "道具#进化道具")...)
	items = append(items, getData(doc, "可交换道具", "道具#可交换道具")...)
	items = append(items, getData(doc, "球果", "道具#可交换道具#球果")...)
	items = append(items, getData(doc, "太晶碎块", "道具#可交换道具#太晶碎块")...)
	items = append(items, getData(doc, "携带物品", "道具#携带物品")...)
	items = append(items, getData(doc, "第二世代", "道具#邮件#第二世代")...)
	items = append(items, getData(doc, "第三世代", "道具#邮件#第三世代")...)
	items = append(items, getData(doc, "第四世代", "道具#邮件#第四世代")...)
	items = append(items, getData(doc, "第五世代", "道具#邮件#第五世代")...)
	items = append(items, getData(doc, "糖果", "道具#糖果")...)
	items = append(items, getData(doc, "护符", "道具#护符")...)
	items = append(items, getData(doc, "材料", "道具#材料")...)
	items = append(items, getData(doc, "精灵球", "精灵球")...)
	items = append(items, getData(doc, "宝物", "宝物")...)
	items = append(items, getData(doc, "贵重道具", "宝物#贵重道具")...)
	items = append(items, getData(doc, "化石", "宝物#化石")...)
	items = append(items, getData(doc, "战斗道具", "战斗道具")...)
	items = append(items, getData(doc, "招式学习器", "招式学习器")...)
	items = append(items, getData(doc, "回复道具", "回复道具")...)
	items = append(items, getData(doc, "训练家使用的Ｚ纯晶", "Z纯晶#训练家使用")...)
	items = append(items, getData(doc, "#宝可梦使用的Ｚ纯晶", "Ｚ纯晶#宝可梦使用的")...)
	items = append(items, getData(doc, "工艺制作", "工艺制作")...)
	items = append(items, getData(doc, "掉落物", "掉落物")...)
	items = append(items, getData(doc, "野餐", "野餐")...)
	items = append(items, getData(doc, "食材", "野餐#食材")...)
	items = append(items, getData(doc, "树果", "树果")...)
	items = append(items, getData(doc, "重要物品", "重要物品")...)
	items = append(items, getData(doc, "洛托姆之力", "洛托姆之力")...)

	var result []interface{}
	for _, value := range items {
		result = append(result, ctx.Output(global.StructToMap(*value)))
	}

	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     result,
	}, nil
}

var imageURLList, nameZhList, nameJaList, nameEnList, descList []string

func extractText(ele *goquery.Selection, offset *int, cache *[]string) string {

	if len(*cache) > 0 {
		text := (*cache)[0]
		*cache = (*cache)[1:]
		return text
	}

	text := strings.TrimSpace(ele.Children().Eq(*offset).Text())

	if val, exists := ele.Children().Eq(*offset).Attr("rowspan"); exists {
		rowCount, _ := strconv.Atoi(val)
		for i := 0; i < rowCount-1; i++ {
			*cache = append(*cache, text)
		}
	}
	*offset++

	return text
}

func parseElement(ele *goquery.Selection, name string) (*Data, error) {
	if name == "第二世代" {
		return &Data{
			NameZh:      strings.TrimSpace(ele.Children().Eq(0).Text()),
			NameJa:      strings.TrimSpace(ele.Children().Eq(1).Text()),
			NameEn:      strings.TrimSpace(ele.Children().Eq(2).Text()),
			Description: strings.TrimSpace(ele.Children().Eq(3).Text()),
		}, nil
	} else if name == "招式学习器" {
		imgEle := ele.Children().Eq(0).Find("img").First()
		imgURL := ""
		if imgEle != nil {
			if val, exists := imgEle.Attr("data-url"); exists {
				imgURL = strings.Replace(val, "//media.52poke.com", "https://s1.52poke.wiki", -1)
			}
		}

		return &Data{
			ImageURL: imgURL,
			NameZh:   strings.TrimSpace(ele.Children().Eq(1).Text()),
			NameJa:   strings.TrimSpace(ele.Children().Eq(2).Text()),
			NameEn:   strings.TrimSpace(ele.Children().Eq(3).Text()),
		}, nil
	}

	offset := 0

	imgURL := ""
	imgEle := ele.Children().Eq(offset).Find("img").First()
	if imgEle != nil {
		if val, exists := imgEle.Attr("data-url"); exists {
			imgURL = strings.Replace(val, "//media.52poke.com", "https://s1.52poke.wiki", -1)
			if val, exists := ele.Children().Eq(offset).Attr("rowspan"); exists {
				rowCount, _ := strconv.Atoi(val)
				for i := 0; i < rowCount-1; i++ {
					imageURLList = append(imageURLList, imgURL)
				}
			}
		}
		offset++
	} else {
		if len(imageURLList) > 0 {
			imgURL = imageURLList[0]
			imageURLList = imageURLList[1:]
		}
	}

	nameZh := extractText(ele, &offset, &nameZhList)
	nameJa := extractText(ele, &offset, &nameJaList)
	nameEn := extractText(ele, &offset, &nameEnList)
	desc := extractText(ele, &offset, &descList)

	data := &Data{
		ImageURL:    imgURL,
		NameZh:      nameZh,
		NameJa:      nameJa,
		NameEn:      nameEn,
		Description: desc,
	}

	return data, nil
}

func getData(doc *goquery.Document, name string, typ string) []*Data {
	var res []*Data

	ele := doc.Find("#" + name).Parent().Next()
	if name == "化石" {
		ele = ele.Next()
	}

	ele.Find("tbody > tr").Each(func(i int, s *goquery.Selection) {
		if i < 1 {
			return
		}

		data, err := parseElement(s, name)

		if err != nil {
			fmt.Println(err)
			return
		}

		data.Type = typ
		res = append(res, data)
	})

	return res
}
