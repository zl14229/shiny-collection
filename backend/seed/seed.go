package seed

import (
	"shiny-collection/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func All(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("seeding database...")

	if err := seedGames(db); err != nil {
		return err
	}
	if err := seedMethods(db); err != nil {
		return err
	}
	if err := seedPokemon(db); err != nil {
		return err
	}

	logger.Info("database seeded successfully")
	return nil
}

func seedGames(db *gorm.DB) error {
	// 先清除旧的游戏数据，用纯 Switch 数据替换
	db.Exec("DELETE FROM games")

	// Nintendo Switch 平台可游玩的 Pokémon 主系列游戏
	games := []model.Game{
		// 去皮去伊 — Gen 7
		{Name: "Pokémon: Let's Go, Pikachu!", NameCN: "宝可梦 Let's Go! 皮卡丘", Generation: 7, Platform: "Nintendo Switch", ShortName: "LGPE", ReleaseYear: 2018},
		{Name: "Pokémon: Let's Go, Eevee!", NameCN: "宝可梦 Let's Go! 伊布", Generation: 7, Platform: "Nintendo Switch", ShortName: "LGPE", ReleaseYear: 2018},
		// 剑盾 — Gen 8
		{Name: "Pokémon Sword", NameCN: "宝可梦 剑", Generation: 8, Platform: "Nintendo Switch", ShortName: "SWSH", ReleaseYear: 2019},
		{Name: "Pokémon Shield", NameCN: "宝可梦 盾", Generation: 8, Platform: "Nintendo Switch", ShortName: "SWSH", ReleaseYear: 2019},
		// 晶灿钻石／明亮珍珠 — Gen 8
		{Name: "Pokémon Brilliant Diamond", NameCN: "宝可梦 晶灿钻石", Generation: 8, Platform: "Nintendo Switch", ShortName: "BDSP", ReleaseYear: 2021},
		{Name: "Pokémon Shining Pearl", NameCN: "宝可梦 明亮珍珠", Generation: 8, Platform: "Nintendo Switch", ShortName: "BDSP", ReleaseYear: 2021},
		// 传说 阿尔宙斯 — Gen 8
		{Name: "Pokémon Legends: Arceus", NameCN: "宝可梦传说 阿尔宙斯", Generation: 8, Platform: "Nintendo Switch", ShortName: "PLA", ReleaseYear: 2022},
		// 朱紫 — Gen 9
		{Name: "Pokémon Scarlet", NameCN: "宝可梦 朱", Generation: 9, Platform: "Nintendo Switch", ShortName: "SV", ReleaseYear: 2022},
		{Name: "Pokémon Violet", NameCN: "宝可梦 紫", Generation: 9, Platform: "Nintendo Switch", ShortName: "SV", ReleaseYear: 2022},
	}

	for _, g := range games {
		db.Where(model.Game{Name: g.Name}).FirstOrCreate(&g)
	}
	return nil
}

func seedMethods(db *gorm.DB) error {
	methods := []model.Method{
		{Name: "Full Odds", NameCN: "纯随机遇敌"},
		{Name: "Masuda Method", NameCN: "异国孵蛋"},
		{Name: "Masuda + Shiny Charm", NameCN: "异国孵蛋+闪符"},
		{Name: "Random Encounter", NameCN: "随机遇敌"},
		{Name: "Soft Reset", NameCN: "软重启"},
		{Name: "Run Away", NameCN: "逃跑流"},
		{Name: "Chain Fishing", NameCN: "连锁钓鱼"},
		{Name: "Pokeradar", NameCN: "宝可追踪(PokéRadar)"},
		{Name: "Friend Safari", NameCN: "朋友狩猎"},
		{Name: "DexNav", NameCN: "图鉴导航(DexNav)"},
		{Name: "Horde Encounter", NameCN: "群聚对战"},
		{Name: "SOS Battle", NameCN: "召唤连锁(SOS)"},
		{Name: "Ultra Wormhole", NameCN: "究极之洞"},
		{Name: "Dynamax Adventure", NameCN: "极巨化冒险"},
		{Name: "Mass Outbreak", NameCN: "大量出现"},
		{Name: "Outbreak + Shiny Charm", NameCN: "大量出现+闪符"},
		{Name: "Mass Outbreak (PLA)", NameCN: "大量出现(阿尔宙斯)"},
		{Name: "MMO (PLA)", NameCN: "大规模大量出现(MMO)"},
		{Name: "Shiny Sandwich", NameCN: "闪力三明治"},
		{Name: "Egg (Breeding)", NameCN: "蛋孵化(普通)"},
		{Name: "Chain Fishing (SV)", NameCN: "Let's Go 连锁"},
		{Name: "Cute Charm Glitch", NameCN: "魅惑身躯漏洞(BDSP)"},
		{Name: "Poke Radar (BDSP)", NameCN: "宝可追踪(BDSP)"},
		{Name: "Other", NameCN: "其他方式"},
	}

	for _, m := range methods {
		db.Where(model.Method{Name: m.Name}).FirstOrCreate(&m)
	}
	return nil
}

func seedPokemon(db *gorm.DB) error {
	// 清除旧数据，重新写入完整 1025 只
	db.Exec("DELETE FROM pokemon")
	for _, p := range AllPokemon {
		db.Where(model.Pokemon{NationalNo: p.NationalNo}).FirstOrCreate(&p)
	}
	return nil
}
