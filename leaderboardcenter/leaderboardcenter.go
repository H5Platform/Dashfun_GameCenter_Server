package leaderboardcenter

import (
	"context"
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/rediscenter"
	"dashfun_gamecenter/usercenter"
	"errors"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"log"
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

var onceLeaderboardCenter sync.Once
var instLeaderboardCenter *LeaderboardCenter

const leaderboardKey = "dashfun_xp_leaderboard"
const leaderboardLoadKey = "dashfun_xp_leaderboard_load"

type LeaderboardData struct {
	Id          string `json:"id"`           //用户ID
	Rank        int64  `json:"rank"`         //用户排名
	Score       int64  `json:"score"`        //用户分数
	UserName    string `json:"username"`     //用户名
	DisplayName string `json:"display_name"` //显示名称
	Avatar      string `json:"avatar"`       //头像
}

type LeaderboardCenter struct {
	sync.RWMutex
	userCache map[string]*data.DashFunUser //user缓存
	loading   bool                         //是否正在从数据库中读取数据建立排行榜中
}

func Get() *LeaderboardCenter {
	onceLeaderboardCenter.Do(func() {
		instLeaderboardCenter = &LeaderboardCenter{}
		instLeaderboardCenter.init()
	})
	return instLeaderboardCenter
}

func (l *LeaderboardCenter) init() {
	l.userCache = make(map[string]*data.DashFunUser)
	events.UserCoinChangedEvents.On(l.onUserCoinChanged)
	events.UserLoginEvents.On(l.onUserLogin)
	if l.needLoad() {
		go l.initRedis()
	} else {
		l.fillLeaderboard()
	}
}

// fillLeaderboard 排行榜不足50人，随机生成50个数据
func (l *LeaderboardCenter) fillLeaderboard() {
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	r, err := rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(50-1)).Result()
	if err != nil {
		log.Fatalf("fillLeaderboard: %s", err.Error())
	}

	maxScore := 0
	for _, z := range r {
		if z.Score > float64(maxScore) {
			maxScore = int(z.Score)
		}
	}

	if maxScore == 0 {
		maxScore = 1000
	}

	if len(r) < 50 {
		members := make([]*redis.Z, 0, 50-len(r))
		for i := len(r); i < 50; i++ {
			uid := "B" + usercenter.Get().RequestUserId()
			members = append(members, &redis.Z{
				Score:  float64((rand.Intn(maxScore*5) + (maxScore / 2)) / 10 * 10),
				Member: uid,
			})
		}
		_, err = rdb.ZAdd(ctx, leaderboardKey, members...).Result()
		if err != nil {
			log.Fatalf("fillLeaderboard: %s", err.Error())
		}
	}

}

func (l *LeaderboardCenter) onUserLogin(user *data.OnlineUser) {
	l.Lock()
	defer l.Unlock()
	_, exist := l.userCache[user.User.Id]
	if exist {
		//update cache only if it's existed in cache
		l.userCache[user.User.Id] = user.User
	}
}

func (l *LeaderboardCenter) initRedis() {
	l.loading = true
	coin, err := dao.GetCoinDao().FindCoinByName("DashFunPoint")
	if err != nil {
		log.Fatalf("find coin failed: %s", err.Error())
	}
	//需要从数据库中读取分数数据，并恢复到redis中
	cursor, err := dao.GetCoinUserDao().GetCoinUserDataCursor(coin.Id, 1000)
	if err != nil {
		zap.S().Fatalf("get point_data cursor failed: %s", err.Error())
	}

	for cursor.Next(context.Background()) {
		point, err := cursor.Data()
		if err != nil {
			zap.S().Errorw("get point_data failed: %s", err.Error())
			continue
		}
		l.updateUserCoinAmount(point.UserId, point.Amount)
	}
	l.fillLeaderboard()
	l.loading = false
}

// GetUserRankAndScore 获取用户的排名
func (l *LeaderboardCenter) GetUserRankAndScore(userId string) (int64, int64, error) {
	if l.loading {
		return 0, 0, nil
	}
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	rank, err := rdb.ZRevRank(ctx, leaderboardKey, userId).Result()
	if err != nil {
		return 0, 0, nil
	}
	score, err := rdb.ZScore(ctx, leaderboardKey, userId).Result()
	if err != nil {
		return 0, 0, nil
	}
	return rank + 1, int64(score), nil
}

// GetTotalUserCount 获取排行榜中的数据总数
func (l *LeaderboardCenter) GetTotalUserCount() (int64, error) {
	if l.loading {
		return 0, nil
	}
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	count, err := rdb.ZCard(ctx, leaderboardKey).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (l *LeaderboardCenter) GetTop(count int) ([]*LeaderboardData, error) {
	if l.loading {
		return make([]*LeaderboardData, 0), nil
	}
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	top := make([]*LeaderboardData, 0)

	for i, z := range r {
		uid := z.Member.(string)
		user, _ := l.getUser(uid)
		if user == nil {
			rdb.ZRem(ctx, leaderboardKey, uid)
			zap.S().Errorw("GetTop100 user failed", "uid", uid, "error", "user not found")
			continue
		}
		top = append(top, &LeaderboardData{
			Id:          uid,
			Rank:        int64(i + 1),
			Score:       int64(z.Score),
			UserName:    user.UserName,
			DisplayName: user.DisplayName,
			Avatar:      user.AvatarUrl,
		})
	}

	return top, nil
}

func (l *LeaderboardCenter) getUser(id string) (*data.DashFunUser, error) {
	l.Lock()
	defer l.Unlock()
	u, ok := l.userCache[id]
	if !ok {
		user, err := usercenter.Get().GetDashFunUser(id)
		if err != nil && !errors.Is(err, apperrors.ErrUserDoesNotExist) {
			return nil, err
		}
		//用户不存在，可能是填充数据，随机做个用户
		if user == nil {
			name := TemplateNames[rand.Intn(len(TemplateNames))]
			user = &data.DashFunUser{
				Id:          id,
				UserName:    name,
				DisplayName: name,
			}
		}
		l.userCache[id] = user
		u = user
	}
	return u, nil
}

func (l *LeaderboardCenter) needLoad() bool {
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//key不存在则新建，返回true，存在则返回false
	result, err := rdb.SetNX(ctx, leaderboardLoadKey, 1, 0).Result()

	if err != nil {
		zap.S().Errorw("redis Exists failed", "key", leaderboardLoadKey, "error", err.Error())
		return false
	}

	return result

}

func (l *LeaderboardCenter) onUserCoinChanged(event events.UserCoinChangedEvent) {
	if event.Coin.Name == "DashFunPoint" {
		l.updateUserCoinAmount(event.UserId, event.UserData.Amount)
	}
}

func (l *LeaderboardCenter) updateUserCoinAmount(userId string, amount int32) {
	if amount <= 0 {
		return
	}
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//将用户的分数存入redis
	_, err := rdb.ZAdd(ctx, leaderboardKey, &redis.Z{
		Score:  float64(amount),
		Member: userId,
	}).Result()

	if err != nil {
		zap.S().Errorw("redis ZAdd failed", "error", err.Error())
	}
}
