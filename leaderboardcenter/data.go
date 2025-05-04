package leaderboardcenter

import "sync"

type ILeaderboardCenter interface {
	GetUserRankAndScore(userId string) (int64, int64, error)
	GetTotalUserCount() (int64, error)
	GetTop(count int) ([]*LeaderboardData, error)
	init()
}

var onceLeaderboardCenter sync.Once
var instLeaderboardCenter ILeaderboardCenter

func Get() ILeaderboardCenter {
	onceLeaderboardCenter.Do(func() {
		instLeaderboardCenter = &LeaderboardCenter{}
		instLeaderboardCenter.init()
	})
	return instLeaderboardCenter
}
