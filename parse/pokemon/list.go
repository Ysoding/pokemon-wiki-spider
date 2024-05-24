package pokemon

import (
	"fmt"

	"github.com/Ysoding/pokemon-wiki-spider/global"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

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
	fmt.Println(1111, ctx.Body)
	items := []interface{}{111, 222}
	return spider.ParseResult{
		Requesrts: make([]*spider.Request, 0),
		Items:     items,
	}, nil
}
