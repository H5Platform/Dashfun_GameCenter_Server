package leaderboardcenter

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/rediscenter"
	"dashfun_gamecenter/spinwheelcenter"
	"dashfun_gamecenter/taskcenter"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"math/rand"
	"sync"
	"time"
)

var TemplateNames = []string{
	"Shadow", "Ghost", "Phantom", "Blaze", "Nova", "Echo", "Storm", "Hunter", "Rebel", "Vortex",
	"KingOfTheGame",
	"SilentKiller",
	"CyberWarrior",
	"NoMercy",
	"DarkLord",
	"FrostByte",
	"MidnightRider",
	"ToxicVenom",
	"QuantumMind",
	"InfinityX",
	"FluffyBear",
	"SweetBunny",
	"ChocoLover",
	"StarrySky",
	"CottonCandy",
	"CutiePie",
	"HappyPanda",
	"CuddleFox",
	"BubbleTeaLove",
	"RainbowDream",
	"0xCyber",
	"AI_Nexus",
	"NeuralNet",
	"CryptoLord",
	"QuantumAI",
	"DataMiner",
	"BlockchainWizard",
	"MetaverseKing",
	"HackerX",
	"GlitchMaster",
	"𝕯𝖆𝖗𝖐𝖊𝖓𝖊𝖉𝕾𝖔𝖚𝖑",
	"☠️Dominator☠️",
	"𝑹𝒆𝒂𝒑𝒆𝒓𝒙𝑿",
	"𝓢𝓱𝓪𝓭𝓸𝔀𝓴𝓲𝓷𝓰",
	"ᎮᏒᏋᏒᎥᏒᎥᏋᏒ",
	"⛓️Underworld⛓️",
	"⭕Omega⭕",
	"⚔️DemonSlayer⚔️",
	"𝕿𝖍𝖊𝕱𝖆𝖑𝖑𝖊𝖓",
	"𝒜𝓃𝓉𝒾-𝒱𝒾𝓇𝓊𝓈",
	"ВечныйОгнь",
	"ЛунныйКот",
	"ЧерныйВолк",
	"КрасныйФеникс",
	"Молния",
	"СибирскийМедведь",
	"КиберВоин",
	"ДождьСлез",
	"Одиночка",
	"СеверныйВетер",
	"Wind Ninja",
	"Shadow Hunter",
	"Fire Spirit",
	"Cyber Wolf",
	"Storm Warrior",
	"Ghost Blade",
	"Thunder King",
	"Ice Phantom",
	"Mystic Raven",
	"Neon Samurai",
	"Dark Angel",
	"Frost Titan",
	"Silent Reaper",
	"Quantum Knight",
	"Eclipse Slayer",
	"Midnight Sniper",
	"Lunar Assassin",
	"Solar Phoenix",
	"Inferno Dragon",
	"Nova Striker",
	"Twilight Demon",
	"Crystal Panther",
	"Sonic Falcon",
	"Toxic Wizard",
	"Cosmic Warrior",
	"Iron Ghost",
	"Storm Breaker",
	"Astral Shadow",
	"Blaze Sniper",
	"Thunder Phantom",
	"Diamond Fox",
	"Turbo Ninja",
	"Chaos Emperor",
	"Venom Swordsman",
	"Void Guardian",
	"Hyper Ghost",
	"Electric Phantom",
	"Silver Ronin",
	"Crimson Hunter",
	"Emerald Assassin",
	"Cursed Reaper",
	"Hellfire Monk",
	"Glitch Sorcerer",
	"Phantom Rider",
	"Titan Slayer",
	"Frost Demon",
	"Mystic Samurai",
	"Arcane Falcon",
	"Celestial Monk",
	"Dark Mirage",
	"Turbo Titan",
	"Toxic Fox",
	"Infernal Sorcerer",
	"Radiant Ghost",
	"Solar Swordsman",
	"Cosmic Ninja",
	"Void Sniper",
	"Blizzard Phantom",
	"Hyper Assassin",
	"Storm Sentinel",
	"Thunder Falcon",
	"Obsidian Striker",
	"Astral Swordsman",
	"Demon Paladin",
	"Cyber Monk",
	"Lunar Reaper",
	"Glacial Phantom",
	"Nova Sentinel",
	"Crystal Sorcerer",
	"Titanium Ronin",
	"Lightning Fox",
	"Chaos Sniper",
	"Frozen Panther",
	"Meteor Warrior",
	"Serpent Assassin",
	"Venom Ronin",
	"Eclipse Phantom",
	"Turbo Mirage",
	"Inferno Reaper",
	"Arcane Titan",
	"Shadow Monk",
	"Solar Striker",
	"Hyper Demon",
	"Diamond Falcon",
	"Titan Assassin",
	"Midnight Paladin",
	"Radiant Slayer",
	"Cosmic Reaper",
	"Astral Mirage",
	"Ice Sorcerer",
	"Obsidian Ninja",
	"Fire Paladin",
	"Thunder Monk",
	"Storm Samurai",
	"Frost Sentinel",
	"Lunar Titan",
	"Electric Warrior",
	"Blizzard Slayer",
	"Void Falcon",
	"Mystic Rider",
	"Теневой Охотник",
	"Лунный Воин",
	"Кибер Волк",
	"Буревестник",
	"Огненный Дракон 🔥",
	"Холодная Тень ❄️",
	"Громовой Король ⚡",
	"Ночной Призрак 👻",
	"Чёрный Самурай",
	"Кристальный Ворон",
	"Бесшумный Убийца",
	"Тёмный Маг 🧙‍♂️",
	"Ледяной Титан ❄️",
	"Кровавый Охотник",
	"Мистический Рыцарь",
	"Астральный Воин",
	"Вечная Буря",
	"Пылающий Феникс 🔥",
	"Грозовой Демон",
	"Ониксовый Дракон 🐉",
	"Бесконечный Ветер",
	"Снежный Ястреб",
	"Серебряный Лис",
	"Огненный Монах 🔥",
	"Чёрный Властелин",
	"Стальной Воин",
	"Электрический Призрак ⚡",
	"Безмолвный Ниндзя",
	"Буря Тьмы",
	"Заклинатель Ветра",
	"Владыка Мороза ❄️",
	"Разрушитель Бурь",
	"Огонь и Пепел 🔥",
	"Скрытая Угроза",
	"Дух Пламени 🔥",
	"Искусственный Разум 🤖",
	"Демон Холода ❄️",
	"Часовой Ночи 🌙",
	"Кровавый Ворон",
	"Солнечный Паладин ☀️",
	"Тёмный Рыцарь",
	"Белый Дракон 🐉",
	"Повелитель Гроз ⚡",
	"Гибельный Клинок",
	"Ледяная Душа ❄️",
	"Смертельная Луна 🌙",
	"Криптовалютный Мастер 💰",
	"Призрачный Странник 👻",
	"Неоновый Самурай",
	"Огненный Шторм 🔥",
}

type LeaderboardBot struct {
	data *data.LeaderboardBotData
	sync.RWMutex
}

// InitScore  生成初始化分数
func (b *LeaderboardBot) InitScore() {
	cfg := config.GetConfig().LeaderboardBotCfg.BotLevels[b.data.Level-1]
	fixedScore := cfg.FixedTaskTop*80/100 + rand.Intn(cfg.FixedTaskTop*20/100+1)
	finalScore := int64(cfg.MinScore + fixedScore)
	b.data.ActiveDate = ""
	b.data.Score = finalScore
	b.data.ActiveDays = 0
	b.data.ActiveTime = rand.Intn(24 * 60 * 60 * 1000)
	b.data.Status = data.LeaderboardBotStatus_Active
}

func (b *LeaderboardBot) Spin() int {
	cfg := config.GetConfig().LeaderboardBotCfg.BotLevels[b.data.Level-1]
	finalScore := 0
	for i := 0; i < cfg.SpinWheelDailyCount; i++ {
		spinWheelHit := spinwheelcenter.Get().RandomSpinWheel()
		if spinWheelHit != nil && spinWheelHit.RewardType == data.SpinWheelReward_DashFunPoint {
			finalScore += spinWheelHit.RewardValue
		}
	}

	finalScore = int(float64(finalScore) * (1 + 0.2*rand.Float64()))
	return finalScore
}

func (b *LeaderboardBot) DoTodayBehaviour() {
	spinScore := b.Spin()
	cfg := config.GetConfig().LeaderboardBotCfg.BotLevels[b.data.Level-1]
	dailyIndex := b.data.ActiveDays
	if dailyIndex >= len(cfg.DailyTop) {
		dailyIndex = len(cfg.DailyTop) - 1
	}
	oldScore := b.data.Score
	dailyScore := cfg.DailyTop[dailyIndex]
	b.data.Score += int64(dailyScore + spinScore)
	b.data.Status = data.LeaderboardBotStatus_DoneToday
	b.data.ActiveDate = time.Now().UTC().Format("20060102")
	b.data.ActiveDays++

	oldRank := b.data.Rank
	if oldRank == 0 {
		oldRank = 9999999
	}

	delta := b.data.Score - oldScore

	rank := b.UploadScore(delta)

	tasks := taskcenter.Get().GetLeaderboardTasks(oldRank, int(rank))

	oldScore = b.data.Score
	scoreChanged := false
	if len(tasks) > 0 {
		for _, task := range tasks {
			for _, r := range task.Rewards {
				if r.RewardType == data.TaskRewardType_DashFunPoint {
					b.data.Score += int64(r.Amount)
					scoreChanged = true
				}
			}
		}
	}

	if scoreChanged {
		delta = b.data.Score - oldScore
		rank = b.UploadScore(delta)
		if rank > 0 {
			b.data.Rank = int(rank)
		}
	}

	dao.GetLeaderboardBotDao().SaveOrUpdate(b.data)
}

// UploadScore 上传分数，并返回最新排名
func (b *LeaderboardBot) UploadScore(delta int64) int64 {
	//存入redis
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//将用户的分数存入redis
	_, err := rdb.ZAdd(ctx, leaderboardKey, &redis.Z{
		Score:  float64(b.data.Score),
		Member: b.data.Id,
	}).Result()

	if err != nil {
		zap.S().Errorw("redis ZAdd failed", "error", err.Error())
	}

	rank, err := rdb.ZRevRank(ctx, leaderboardKey, b.data.Id).Result()
	if err != nil {
		zap.S().Errorw("redis ZRevRank failed", "error", err.Error())
		return 0
	}

	return rank + 1
}
