package leaderboardcenter

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/utils"
	"log"
	"math/rand"
	"time"
)

type LeaderboardBotRpc struct {
	LeaderboardBot
}

// UploadScore 上传分数，并返回最新排名
func (b *LeaderboardBotRpc) UploadScore() int64 {
	return 0
}

type BotManager struct {
	bots utils.Map[string, *LeaderboardBot]
}

func (bm *BotManager) LoadAllBots() {
	bm.bots = utils.NewUnsafeMap[string, *LeaderboardBot]()

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
		bm.bots.Set(b.Id, bot)
	}

	if len(bots) < LeaderboardMinCount {
		//填满bot数据
		for i := len(bots); i < LeaderboardMinCount; i++ {
			botLevel := bm.randomBotLevel()
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
			bm.bots.Set(nbd.Id, bot)
		}
	}
}

func (bm *BotManager) randomBotLevel() int {
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

func (bm *BotManager) botBehaviour() {
	now := time.Now().UTC()
	midnight := now.Truncate(24 * time.Hour)
	t := int(now.Sub(midnight).Milliseconds())

	today := time.Now().UTC().Format("20060102")

	bm.bots.Range(func(key string, bot *LeaderboardBot) bool {
		bd := bot.data
		bot.Lock()
		defer bot.Unlock()

		//如果今天已经完成了任务，且不是今天的日期，则需要重置状态
		if bd.Status == data.LeaderboardBotStatus_DoneToday && bd.ActiveDate != today {
			bd.Status = data.LeaderboardBotStatus_Active
		}

		if bd.Status == data.LeaderboardBotStatus_Active && t > bd.ActiveTime {
			bot.DoTodayBehaviour()
		}
		return true
	})

}

func (bm *BotManager) Run() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bm.botBehaviour()
			}
		}
	}()
}
