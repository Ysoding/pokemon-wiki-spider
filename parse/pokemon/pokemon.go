package pokemon

import (
	"github.com/Ysoding/pokemon-wiki-spider/parse/pokemon/ability"
	"github.com/Ysoding/pokemon-wiki-spider/parse/pokemon/move"
	"github.com/Ysoding/pokemon-wiki-spider/parse/pokemon/nature"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

var Tasks = []*spider.Task{
	PokemonListTask,
	ability.PokemonAbilityListTask,
	ability.AbilityListTask,
	nature.PokemonNatureListTask,
	move.MoveListTask,
}
