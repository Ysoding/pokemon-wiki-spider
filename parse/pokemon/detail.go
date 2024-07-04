package pokemon

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
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

type PokemonDetailData struct {
	Index               int
	NameZh              string
	ImgURL              string // 图片链接
	Type                string // 属性
	Category            string // 分类
	Ability             string // 特性
	Height              string // 身高
	Weight              string // 体重
	BodyStyle           string // 体形
	CatchRate           string // 捕获率
	GenderRatio         string // 性别比例
	EggGroup1           string // 每一生蛋分组
	EggGroup2           string // 第二生蛋分组
	HatchTime           string // 孵化时间
	EffortValue         string // 基础点数
	BaseStat            BaseStat
	LearnableMovesList  []LearnableMove   // 可学会的招式
	UsableMoveTutorList []UsableMoveTutor // 能使用的招式学习器
	EggMoveList         []EggMove         // 蛋招式
}

type BaseStat struct {
	HP        int
	Attack    int
	Defense   int
	SpAttack  int
	SpDefense int
	Speed     int
	Total     int
	Average   float32
}

type LearnableMove struct {
	Level    string
	Move     string
	Type     string
	Category string
	Power    string
	Accuracy string
	PP       string
}

type UsableMoveTutor struct {
	ImgURL           string
	TechnicalMachine string
	Move             string
	Type             string
	Category         string
	Power            string
	Accuracy         string
	PP               string
}

type EggMove struct {
	Parent   string
	Move     string
	Type     string
	Category string
	Power    string
	Accuracy string
	PP       string
}

var PokemonDetailTask = &spider.Task{
	Options: spider.Options{
		Name:     global.PokemonDetailTaskName,
		Cookie:   "",
		MaxDepth: 5,
		WaitTime: 3,
		Limit: limiter.Multi(
			rate.NewLimiter(limiter.Per(1, 1*time.Second), 1),
		),
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
		req := &spider.Request{
			URL:      fmt.Sprintf("https://wiki.52poke.com/zh-hans/%s", d.NameZh),
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

func parsePokemonDetail(ctx *spider.Context) (spider.ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		return spider.ParseResult{}, err
	}

	index, _ := ctx.Req.TempData.Get("index").(int)
	indexStr := fmt.Sprintf("%03d", index)
	nameZh, _ := ctx.Req.TempData.Get("nameZh").(string)

	table := doc.Find("#mw-content-text > .mw-parser-output > table").Eq(1)

	imgURL := ""
	if img := table.Find(fmt.Sprintf("img[alt^=%s]", indexStr)).First(); img.Length() != 0 {
		imgURL = img.AttrOr("data-url", "")
		imgURL = strings.Replace(imgURL, "//media.52poke.com", "https://s1.52poke.wiki", 1)
	}

	typeStrs := table.Find("[title=属性]").Parent().Next().Find("span[class=type-box-9-text]").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})
	typeStr := formatStr(strings.Join(typeStrs, ","))

	category := formatStr(table.Find("[title=分类]").Parent().Next().Text())

	tmp := table.Find("[title=特性]").Parent().Next().Find("td")
	abilityList := tmp.Eq(0).Find("a").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})
	if tmp.Length() > 1 {
		abilityHide := tmp.Eq(1).Find("a").Text() + "（隐藏特性）"
		abilityList = append(abilityList, abilityHide)
	}
	ability := formatStr(strings.Join(abilityList, ","))

	tmp = table.Find("[title=宝可梦列表（按身高和体重排序）]")
	height := formatStr(tmp.Eq(0).Parent().Next().Text())
	weight := formatStr(tmp.Eq(1).Parent().Next().Text())

	tmp = table.Find("[title=宝可梦列表（按体形分类）]").Parent().Next()
	bodyStyle := "无"
	if tmp.Find("img").Length() != 0 {
		bodyStyle = strings.Replace(tmp.Find("img").AttrOr("data-url", ""), "//media.52poke.com", "https://s1.52poke.wiki", 1)
	}

	catchRate := formatStr(table.Find("[title=捕获率]").Parent().Next().Find("span.explain").Text())

	genderRatioList := table.Find("[title=宝可梦列表（按性别比例分类）]").Parent().Next().Find("span").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})
	genderRatio := "无性别"
	if len(genderRatioList) > 0 {
		genderRatio = formatStr(strings.Join(genderRatioList, ","))
	}

	eggGroupList := table.Find("[title=宝可梦培育]").Parent().Next().Find("td").Eq(0).Find("a").Map(func(i int, s *goquery.Selection) string {
		return strings.ReplaceAll(strings.TrimSpace(s.AttrOr("title", "")), "（.*）", "")
	})
	eggGroup1 := ""
	if len(eggGroupList) >= 1 {
		eggGroup1 = eggGroupList[0]
	}
	eggGroup2 := ""
	if len(eggGroupList) >= 2 {
		eggGroup2 = eggGroupList[1]
	}

	hatchTime := formatStr(table.Find("[title=宝可梦培育]").Parent().Next().Find("td").Eq(1).Text())

	effortValueList := table.Find("[title=基础点数]").Parent().Next().Find("tr").Find("td").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})
	effortValue := formatStr(strings.Join(effortValueList, ","))

	var result []interface{}
	result = append(result, ctx.Output(global.StructToMap(&PokemonDetailData{
		Index:               index,
		NameZh:              nameZh,
		ImgURL:              imgURL,
		Type:                typeStr,
		Category:            category,
		Ability:             ability,
		Height:              height,
		Weight:              weight,
		BodyStyle:           bodyStyle,
		CatchRate:           catchRate,
		GenderRatio:         genderRatio,
		EggGroup1:           eggGroup1,
		EggGroup2:           eggGroup2,
		HatchTime:           hatchTime,
		EffortValue:         effortValue,
		BaseStat:            parseBaseStat(doc),
		LearnableMovesList:  parseLearnableMovesList(doc),
		UsableMoveTutorList: parseUsableMoveTutorList(doc),
		EggMoveList:         parseEggMoveList(doc, nameZh),
	})))

	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     result,
	}, nil
}

func parseBaseStat(doc *goquery.Document) BaseStat {
	baseStatSpan := doc.Find("#种族值").First()
	baseStatTable := baseStatSpan.Parent().NextAllFiltered("table").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.HasClass("roundy")
	}).First()

	hp, _ := strconv.Atoi(baseStatTable.Find("tr.bgl-HP span[style*='float:right']").Text())
	attack, _ := strconv.Atoi(baseStatTable.Find("tr.bgl-攻击 span[style*='float:right']").Text())
	defense, _ := strconv.Atoi(baseStatTable.Find("tr.bgl-防御 span[style*='float:right']").Text())
	spAttack, _ := strconv.Atoi(baseStatTable.Find("tr.bgl-特攻 span[style*='float:right']").Text())
	spDefense, _ := strconv.Atoi(baseStatTable.Find("tr.bgl-特防 span[style*='float:right']").Text())
	speed, _ := strconv.Atoi(baseStatTable.Find("tr.bgl-速度 span[style*='float:right']").Text())

	total := hp + attack + defense + spAttack + spDefense + speed
	average := float32(total) / 6.0

	return BaseStat{
		HP:        hp,
		Attack:    attack,
		Defense:   defense,
		SpAttack:  spAttack,
		SpDefense: spDefense,
		Speed:     spAttack,
		Total:     total,
		Average:   average,
	}
}

func parseLearnableMovesList(doc *goquery.Document) []LearnableMove {
	res := make([]LearnableMove, 0)

	tmpEle := doc.Find("#可学会的招式").First()
	if tmpEle.Length() == 0 {
		return res
	}

	tmpEle = tmpEle.Parent().Next()
	if tmpEle.Length() == 0 || tmpEle.Nodes[0].Data != "table" {
		return res
	}

	tmpEle.Find("tbody > tr.bgwhite").Each(func(i int, s *goquery.Selection) {
		tdList := s.Find("td:not(.hide)")

		level := formatStr(tdList.Eq(0).Text())
		move := formatStr(tdList.Eq(1).Find("a").Text())
		typ := formatStr(tdList.Eq(2).Find("a").Text())
		category := formatStr(tdList.Eq(3).Find("a").Text())
		power := formatStr(tdList.Eq(4).Text())
		accuracy := formatStr(tdList.Eq(5).Text())
		pp := formatStr(tdList.Eq(6).Text())

		res = append(res, LearnableMove{
			Level:    level,
			Move:     move,
			Type:     typ,
			Category: category,
			Power:    power,
			Accuracy: accuracy,
			PP:       pp,
		})
	})

	return res
}

func parseUsableMoveTutorList(doc *goquery.Document) []UsableMoveTutor {
	res := make([]UsableMoveTutor, 0)

	tmpEle := doc.Find("#能使用的招式学习器").First()
	if tmpEle.Length() == 0 {
		tmpEle = doc.Find("#能使用的招式学习器和招式记录").First()
	}
	if tmpEle.Length() == 0 {
		return res
	}

	tmpEle = tmpEle.Parent().Next()
	if tmpEle.Length() == 0 || tmpEle.Nodes[0].Data != "table" {
		return res
	}

	tmpEle.Find("tbody > tr.bgwhite").Each(func(i int, s *goquery.Selection) {
		imgURL := strings.Replace(s.Find("img").AttrOr("data-url", ""), "//media.52poke.com", "https://s1.52poke.wiki", 1)
		technicalMachine := formatStr(s.Children().Eq(1).Text())
		move := formatStr(s.Children().Eq(2).Find("a").Text())
		typ := formatStr(s.Children().Eq(3).Text())
		category := formatStr(s.Children().Eq(4).Text())
		power := formatStr(s.Children().Eq(5).Text())
		accuracy := formatStr(s.Children().Eq(6).Text())
		pp := formatStr(s.Children().Eq(7).Text())

		res = append(res, UsableMoveTutor{
			ImgURL:           imgURL,
			TechnicalMachine: technicalMachine,
			Move:             move,
			Type:             typ,
			Category:         category,
			Power:            power,
			Accuracy:         accuracy,
			PP:               pp,
		})
	})
	return res
}

func parseEggMoveList(doc *goquery.Document, nameZh string) []EggMove {
	res := make([]EggMove, 0)
	tmpEle := doc.Find("#蛋招式").First()
	if tmpEle.Length() == 0 {
		return res
	}

	tmpEle.Parent().Next().Find("tbody > tr.bgwhite").Each(func(i int, s *goquery.Selection) {
		parentList := s.Children().Eq(0).Find("span, a").Map(func(i int, e *goquery.Selection) string {
			if e.Is("a") {
				if e.HasClass("selflink") {
					return nameZh
				} else {
					title := e.AttrOr("title", "")
					if title == "模仿香草" {
						return ""
					}
					return title
				}
			} else if e.Is("span") {
				msp := e.AttrOr("data-msp", "")
				if msp != "" {
					parts := strings.Split(msp, ",")
					var names []string
					for _, part := range parts {
						if strings.Contains(part, "\\") {
							names = append(names, part[strings.Index(part, "\\")+1:])
						}
					}
					return strings.Join(names, ",")
				}
			}
			return ""
		})

		filteredParentList := filterStrings(parentList, func(s string) bool {
			return s != ""
		})

		parent := strings.Join(filteredParentList, ", ")

		move := formatStr(s.Children().Eq(1).Find("a").Text())
		typ := formatStr(s.Children().Eq(2).Text())
		category := formatStr(s.Children().Eq(3).Text())
		power := formatStr(s.Children().Eq(4).Text())
		accuracy := formatStr(s.Children().Eq(5).Text())
		pp := formatStr(s.Children().Eq(6).Text())

		res = append(res, EggMove{
			Parent:   parent,
			Move:     move,
			Type:     typ,
			Category: category,
			Power:    power,
			Accuracy: accuracy,
			PP:       pp,
		})
	})
	return res
}

func filterStrings(strings []string, predicate func(string) bool) []string {
	var filtered []string
	for _, s := range strings {
		if predicate(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func formatStr(str string) string {
	return strings.ReplaceAll(strings.TrimSpace(str), "\n", "")
}
