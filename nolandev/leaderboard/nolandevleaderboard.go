package nolandevleaderboard

import (
	"context"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/nolandev/nolandata"
	"dashfun_gamecenter/rediscenter"
	"dashfun_gamecenter/usercenter"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const leaderboardKey = "nolan_dp_leaderboard"
const leaderboardLoadKey = "nolan_dp_leaderboard_load"
const LeaderboardMinCount = 1234

var onceLeaderboardCenter sync.Once
var instLeaderboardCenter *NolanDevLeaderboardCenter

type NolanDevLeaderboard struct {
	Id          string `json:"id"`           //用户ID
	Rank        int64  `json:"rank"`         //用户排名
	Score       int64  `json:"score"`        //用户分数
	DisplayName string `json:"display_name"` //显示名称
}

type NolanDevLeaderboardCenter struct {
	sync.RWMutex
	userCache map[string]*data.DashFunUser      //user缓存
	loading   bool                              //是否正在从数据库中读取数据建立排行榜中
	bots      map[string]*nolandata.NolanDevBot //所有的机器人
}

func Get() *NolanDevLeaderboardCenter {
	onceLeaderboardCenter.Do(func() {
		instLeaderboardCenter = &NolanDevLeaderboardCenter{}
		instLeaderboardCenter.init()
	})
	return instLeaderboardCenter
}

func (l *NolanDevLeaderboardCenter) init() {
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

}

// loadAllBots 加载所有的机器人数据
func (l *NolanDevLeaderboardCenter) loadAllBots() {
	//从数据库中读取bot
	l.bots = make(map[string]*nolandata.NolanDevBot)

	botDao := dao.GetNolanDevLeaderboardBotDao()
	botData, err := botDao.LoadAllBots()
	if err != nil {
		log.Fatalf("loadAllBots err: %v", err)
	}

	bots := make([]*nolandata.NolanDevBot, 0)

	for _, b := range botData {
		bot := &nolandata.NolanDevBot{
			Data: b,
		}
		bots = append(bots, bot)
		l.bots[b.Id] = bot
	}

	//needInitPost := len(bots) == 0

	if len(bots) < LeaderboardMinCount {
		//需要补充bot
		botCount := LeaderboardMinCount - len(bots)
		for i := 0; i < botCount; i++ {
			bot := nolandata.RandomBot()
			botDao.SaveOrUpdate(bot.Data)
			bots = append(bots, bot)
			l.bots[bot.Data.Id] = bot
		}
	}

	members := make([]*redis.Z, 0, len(bots))
	for _, bot := range bots {
		members = append(members, &redis.Z{
			Score:  float64(bot.Data.Score),
			Member: bot.Data.Id,
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

	//如果数据库中没有bot，则表示初始化，需要在生成完bot后，抽150个bot发post，初始化post数据

}

func (l *NolanDevLeaderboardCenter) botBehaviour() {
	now := time.Now().UnixMilli()

	botDao := dao.GetNolanDevLeaderboardBotDao()
	for _, bot := range l.bots {
		bd := bot.Data
		bot.Lock()

		////如果今天已经完成了任务，且不是今天的日期，则需要重置状态
		//if bd.Status == BotStatus_DoneToday && bd.ActiveDate != today {
		//	bd.Status = data.LeaderboardBotStatus_Active
		//}

		if bd.Status == data.BotStatus_Active && now > bd.ActiveTime {
			bot.DoTodayBehaviour()
			rank := l.UploadScore(bot.Data)
			bot.Data.Rank = rank
			botDao.SaveOrUpdate(bot.Data)
		}
		bot.Unlock()
	}
}

// UploadScore 上传分数，并返回最新排名
func (l *NolanDevLeaderboardCenter) UploadScore(bot *data.NolanBotData) int64 {
	//存入redis
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//将用户的分数存入redis
	_, err := rdb.ZAdd(ctx, leaderboardKey, &redis.Z{
		Score:  float64(bot.Score),
		Member: bot.Id,
	}).Result()

	if err != nil {
		zap.S().Errorw("redis ZAdd failed", "error", err.Error())
	}

	rank, err := rdb.ZRevRank(ctx, leaderboardKey, bot.Id).Result()
	if err != nil {
		zap.S().Errorw("redis ZRevRank failed", "error", err.Error())
		return 0
	}

	return rank + 1
}
func (l *NolanDevLeaderboardCenter) needLoad() bool {
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

func (l *NolanDevLeaderboardCenter) initRedis() {
	l.loading = true
	coin, err := dao.GetCoinDao().FindCoinByName("NolanDevPoint")
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
	l.loading = false
}

// fillLeaderboard 随机生成排行榜数据
func (l *NolanDevLeaderboardCenter) fillLeaderboard() {
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	r, err := rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(LeaderboardMinCount-1)).Result()
	if err != nil {
		log.Fatalf("fillLeaderboard: %s", err.Error())
	}

	maxScore := 50000

	if len(r) < LeaderboardMinCount {
		members := make([]*redis.Z, 0, LeaderboardMinCount-len(r))
		for i := len(r); i < LeaderboardMinCount; i++ {
			uid := "BT" + usercenter.Get().RequestUserId()
			members = append(members, &redis.Z{
				Score:  float64((rand.Intn(maxScore*2) + (maxScore / 2)) / 500 * 500),
				Member: uid,
			})
		}
		_, err = rdb.ZAdd(ctx, leaderboardKey, members...).Result()
		if err != nil {
			log.Fatalf("fillLeaderboard: %s", err.Error())
		}
	}
}

func (l *NolanDevLeaderboardCenter) onUserCoinChanged(event events.UserCoinChangedEvent) {
	if event.Coin.Name == "NolanDevPoint" {
		l.updateUserCoinAmount(event.UserId, event.UserData.Amount)
	}
}

func (l *NolanDevLeaderboardCenter) updateUserCoinAmount(userId string, amount int32) {
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

func (l *NolanDevLeaderboardCenter) onUserLogin(user *data.OnlineUser) {
	l.Lock()
	defer l.Unlock()
	_, exist := l.userCache[user.User.Id]
	if exist {
		//update cache only if it's existed in cache
		l.userCache[user.User.Id] = user.User
	}
}
func (l *NolanDevLeaderboardCenter) GetUserRankAndScore(userId string) (int64, int64, error) {
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

func (l *NolanDevLeaderboardCenter) GetTop(count int) ([]*NolanDevLeaderboard, error) {
	if l.loading {
		return make([]*NolanDevLeaderboard, 0), nil
	}
	rdb := rediscenter.Get()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	top := make([]*NolanDevLeaderboard, 0)

	for i, z := range r {
		uid := z.Member.(string)
		user, _ := l.getUser(uid)
		if user == nil {
			rdb.ZRem(ctx, leaderboardKey, uid)
			zap.S().Errorw("GetTop100 user failed", "uid", uid, "error", "user not found")
			continue
		}
		top = append(top, &NolanDevLeaderboard{
			Id:          uid,
			Rank:        int64(i + 1),
			Score:       int64(z.Score),
			DisplayName: user.DisplayName,
		})
	}

	return top, nil
}

func (l *NolanDevLeaderboardCenter) getUser(id string) (*data.DashFunUser, error) {
	l.Lock()
	defer l.Unlock()
	u, ok := l.userCache[id]
	var user *data.DashFunUser
	if !ok {
		if strings.HasPrefix(id, "fvb") {
			//bot user
			bot := l.bots[id]
			user = &data.DashFunUser{
				Id:          id,
				UserName:    bot.Data.Name,
				DisplayName: bot.Data.Name,
			}
		} else {
			user, _ = usercenter.Get().GetDashFunUser(id)
		}
		//user, _ := usercenter.Get().GetDashFunUser(id)
		//if user == nil {
		//	botUser := l.bots[id]
		//	name := ""
		//	if botUser == nil {
		//		name = TemplateNames[rand.Intn(len(TemplateNames))]
		//	} else {
		//		if botUser.data.Name == "" {
		//			botUser.Lock()
		//			botUser.data.Name = TemplateNames[rand.Intn(len(TemplateNames))]
		//			dao.GetLeaderboardBotDao().SaveOrUpdate(botUser.data)
		//			botUser.Unlock()
		//		}
		//		name = botUser.data.Name
		//	}
		//	user = &data.DashFunUser{
		//		Id:          id,
		//		UserName:    name,
		//		DisplayName: name,
		//	}
		//}
		l.userCache[id] = user
		u = user
	}
	return u, nil
}
