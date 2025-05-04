package leaderboardcenter

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/rediscenter"
	"dashfun_gamecenter/usercenter"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"sync"
	"time"
)

const leaderboardKey = "dashfun_xp_leaderboard"
const leaderboardLoadKey = "dashfun_xp_leaderboard_load"
const LeaderboardMinCount = 1666

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
	bots      map[string]*LeaderboardBot   //所有的机器人
}

func (l *LeaderboardCenter) init() {
	l.userCache = make(map[string]*data.DashFunUser)
	events.UserCoinChangedEvents.On(l.onUserCoinChanged)
	events.UserLoginEvents.On(l.onUserLogin)
	l.loadAllBots()
	if l.needLoad() {
		go l.initRedis()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.botBehaviour()
			}
		}
	}()

	//else {
	//	l.fillLeaderboard()
	//}
}

func (l *LeaderboardCenter) botBehaviour() {
	now := time.Now().UTC()
	midnight := now.Truncate(24 * time.Hour)
	t := int(now.Sub(midnight).Milliseconds())

	today := time.Now().UTC().Format("20060102")

	for _, bot := range l.bots {
		bd := bot.data
		bot.Lock()

		//如果今天已经完成了任务，且不是今天的日期，则需要重置状态
		if bd.Status == data.LeaderboardBotStatus_DoneToday && bd.ActiveDate != today {
			bd.Status = data.LeaderboardBotStatus_Active
		}

		if bd.Status == data.LeaderboardBotStatus_Active && t > bd.ActiveTime {
			bot.DoTodayBehaviour()
		}
		bot.Unlock()
	}
}

func (l *LeaderboardCenter) randomBotLevel() int {
	cfgs := config.GetConfig().LeaderboardBotCfg.BotLevels
	totalWeight := 0
	for _, cfg := range cfgs {
		totalWeight += cfg.Weight
	}

	randValue := rand.Intn(totalWeight)
	for _, cfg := range cfgs {
		if randValue < cfg.Weight {
			return cfg.Level // 等级从1开始
		}
		randValue -= cfg.Weight
	}
	return 1 // 默认返回最低等级
}

func (l *LeaderboardCenter) loadAllBots() {
	l.bots = make(map[string]*LeaderboardBot)

	botDao := dao.GetLeaderboardBotDao()
	botData, err := botDao.LoadAllBots()
	if err != nil {
		log.Fatalf("loadAllBots err: %v", err)
	}

	bots := make([]*LeaderboardBot, 0)

	for _, b := range botData {
		bot := &LeaderboardBot{
			data: b,
		}
		bots = append(bots, bot)
		l.bots[b.Id] = bot
	}

	if len(bots) < LeaderboardMinCount {
		//填满bot数据
		for i := len(bots); i < LeaderboardMinCount; i++ {
			botLevel := l.randomBotLevel()
			nbd := &data.LeaderboardBotData{
				Id:    "b" + usercenter.Get().RequestUserId(),
				Name:  "", //name先空着，需要的时候再随机
				Level: botLevel,
			}
			bot := &LeaderboardBot{
				data: nbd,
			}
			bot.InitScore()
			botDao.SaveOrUpdate(bot.data)
			bots = append(bots, bot)
			l.bots[nbd.Id] = bot
		}
	}

	members := make([]*redis.Z, 0, len(bots))
	for _, bot := range bots {
		members = append(members, &redis.Z{
			Score:  float64(bot.data.Score),
			Member: bot.data.Id,
		})
	}

	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	//将所有的bot数据存入redis
	_, err = rdb.ZAdd(ctx, leaderboardKey, members...).Result()
	if err != nil {
		log.Fatalf("add bots to redis error: %s", err.Error())
	}
}

// fillLeaderboard 随机生成排行榜数据
func (l *LeaderboardCenter) fillLeaderboard() {
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	r, err := rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(LeaderboardMinCount-1)).Result()
	if err != nil {
		log.Fatalf("fillLeaderboard: %s", err.Error())
	}

	maxScore := 5000

	if len(r) < LeaderboardMinCount {
		members := make([]*redis.Z, 0, LeaderboardMinCount-len(r))
		for i := len(r); i < LeaderboardMinCount; i++ {
			uid := "B" + usercenter.Get().RequestUserId()
			members = append(members, &redis.Z{
				Score:  float64((rand.Intn(maxScore*2) + (maxScore / 2)) / 100 * 100),
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
		user, _ := usercenter.Get().GetDashFunUser(id)
		//if err != nil && !errors.Is(err, apperrors.ErrUserDoesNotExist) {
		//	return nil, err
		//}
		//用户不存在，可能是填充数据，随机做个用户
		if user == nil {
			botUser := l.bots[id]
			name := ""
			if botUser == nil {
				name = TemplateNames[rand.Intn(len(TemplateNames))]
			} else {
				if botUser.data.Name == "" {
					botUser.Lock()
					botUser.data.Name = TemplateNames[rand.Intn(len(TemplateNames))]
					dao.GetLeaderboardBotDao().SaveOrUpdate(botUser.data)
					botUser.Unlock()
				}
				name = botUser.data.Name
			}
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

	rank, err := rdb.ZRevRank(ctx, leaderboardKey, userId).Result()
	if err != nil {
		zap.S().Errorw("redis ZRevRank failed", "error", err.Error())
		return
	}

	events.UserLeaderboardEvents.Emit(&events.UserLeaderboardEvent{
		Id:     "",
		UserId: userId,
		Rank:   rank + 1,
		Score:  float64(amount),
	})

}
