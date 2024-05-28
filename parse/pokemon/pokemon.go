package pokemon

import (
	"github.com/Ysoding/pokemon-wiki-spider/parse/pokemon/item"
	"github.com/Ysoding/pokemon-wiki-spider/spider"
)

var Tasks = []*spider.Task{
	// PokemonListTask,
	// ability.PokemonAbilityListTask,
	// ability.AbilityListTask,
	// nature.PokemonNatureListTask,
	// move.MoveListTask,
	item.ItemListTask,
}
