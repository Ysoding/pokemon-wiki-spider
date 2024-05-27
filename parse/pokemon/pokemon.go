package pokemon

import (
	"github.com/Ysoding/pokemon-wiki-spider/parse/pokemon/ability"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

var Tasks = []*spider.Task{
	PokemonListTask,
	ability.PokemonAbilityListTask,
}
