package leaderboardcenter

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/nacoscenter"
	"dashfun_gamecenter/utils"
	"go.uber.org/zap"
	"time"
)

type LeaderboardCenterRpc struct {
	leaderboardRpc *nacoscenter.LeaderboardRpc
	name2defs      utils.Map[string, *data.LeaderboardDefine]
	id2defs        utils.Map[string, *data.LeaderboardDefine]
	userCache      *utils.GenericCache[*data.DashFunUser]
	bots           utils.Map[string, *LeaderboardBot]
}

func (l *LeaderboardCenterRpc) GetUserRankAndScore(userId string) (int64, int64, error) {
	panic("implement me")
}

func (l *LeaderboardCenterRpc) GetTotalUserCount() (int64, error) {
	panic("implement me")
}

func (l *LeaderboardCenterRpc) GetTop(count int) ([]*LeaderboardData, error) {
	panic("implement me")
}

func (l *LeaderboardCenterRpc) init() {
	l.leaderboardRpc = nacoscenter.GetLeaderboardRpc()

	l.name2defs = utils.NewUnsafeMap[string, *data.LeaderboardDefine]()
	l.id2defs = utils.NewUnsafeMap[string, *data.LeaderboardDefine]()

	l.userCache = utils.NewCache[*data.DashFunUser](30 * time.Minute)

	events.UserCoinChangedEvents.On(l.onUserCoinChanged)
	events.UserLoginEvents.On(l.onUserLogin)

	cfg := config.GetConfig().LeaderboardCfg

	for _, c := range cfg {
		def, isNew, err := l.leaderboardRpc.CreateLeaderboardNX(c.Name, c.GameId, c.PeriodType, string(c.ScoreType))
		if err != nil {
			zap.S().Errorw("LeaderboardCenterRpc init", "cfg", cfg, "err", err)
			continue
		}

		l.name2defs.Set(def.Name, def)
		l.id2defs.Set(def.Id, def)

		if isNew {
			//init leaderboard

		}
	}

	//l.loadAllBots()
}

func (l *LeaderboardCenterRpc) onUserCoinChanged(event events.UserCoinChangedEvent) {
	if event.Coin.Name == "DashFunPoint" {
	}
}

func (l *LeaderboardCenterRpc) onUserLogin(user *data.OnlineUser) {
	_, exist := l.userCache.Get(user.User.Id)
	if exist {
		//update cache only if it's existed in cache
		l.userCache.Set(user.User.Id, user.User)
	}
}
