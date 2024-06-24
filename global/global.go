package global

var (
	EnableMongoDB            = true
	DefaultMongoDatabaseName = "pokemon"
	DefaultBatchCount        = 100
	PokemonListURL           = "https://wiki.52poke.com/wiki/宝可梦列表（按全国图鉴编号）"
	PokemonAbilityListURL    = "https://wiki.52poke.com/zh-hans/特性列表（按全国图鉴编号）"
	AbilityListURL           = "https://wiki.52poke.com/zh-hans/特性列表"
	PokemonNatureListURL     = "https://wiki.52poke.com/zh-hans/性格"
	PokemonMoveListURL       = "https://wiki.52poke.com/zh-hans/招式列表"
	PokemonItemListURL       = "https://wiki.52poke.com/zh-hans/道具列表"

	DefaultWorkerCount = 16

	LocationNameList = []string{
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
)
