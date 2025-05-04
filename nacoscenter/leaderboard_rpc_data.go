package nacoscenter

import (
	"dashfun_gamecenter/datasource/data"
	"github.com/dashfun_web3/api_proto/gen/common"
)

func PbToLeaderboardDefine(pb *common.LeaderboardDataPb) *data.LeaderboardDefine {
	return &data.LeaderboardDefine{
		Id:         pb.LeaderboardId,
		Name:       pb.Name,
		GameId:     pb.GameId,
		PeriodType: data.LeaderboardPeriodType(pb.PeriodType),
		ScoreType:  pb.ScoreType,
		InitTime:   pb.InitTime,
		ResetTime:  pb.ResetTime,
		Status:     data.LeaderboardStatus(pb.Status),
	}
}
func PbToLeaderboardRankData(pb *common.LeaderboardRankDataPb) (*data.LeaderboardRankData, string) {
	return &data.LeaderboardRankData{
		UserId: pb.UserId,
		Rank:   pb.Rank,
		Score:  pb.Score,
	}, pb.LeaderboardId
}
