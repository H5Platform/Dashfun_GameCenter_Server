package nacoscenter

import (
	"context"
	"dashfun_gamecenter/datasource/data"
	v1 "github.com/dashfun_web3/api_proto/gen/leaderboardservice/v1"
	"go.uber.org/zap"
	"time"
)

type LeaderboardRpc struct {
}

func GetLeaderboardRpc() *LeaderboardRpc {
	_, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("GetLeaderboardRpc", "err", err)
		return nil
	}
	return &LeaderboardRpc{}
}

func (l *LeaderboardRpc) CreateLeaderboardNX(name, gameId string, periodType data.LeaderboardPeriodType, scoreType string) (*data.LeaderboardDefine, bool, error) {
	leaderboardServiceClient, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("CreateLeaderboardNX", "err", err)
		return nil, false, err
	}
	if leaderboardServiceClient == nil {
		zap.S().Errorw("CreateLeaderboardNX", "err", "leaderboard service client is nil")
		return nil, false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := leaderboardServiceClient.CreateLeaderboardNX(ctx, &v1.CreateLeaderboardNXRequest{
		Name:       name,
		GameId:     gameId,
		PeriodType: int32(periodType),
		ScoreType:  scoreType,
	})
	if err != nil {
		zap.S().Errorw("CreateLeaderboardNX", "err", err)
		return nil, false, err
	}
	zap.S().Infow("CreateLeaderboardNX", "resp", resp)
	d := PbToLeaderboardDefine(resp.Leaderboard)
	return d, resp.IsNew, nil
}

func (l *LeaderboardRpc) GetLeaderboardById(leaderboardId string) (*data.LeaderboardDefine, error) {
	c, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("GetLeaderboardById", "err", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.GetLeaderboard(ctx, &v1.GetLeaderboardRequest{
		LeaderboardId: leaderboardId,
	})
	if err != nil {
		return nil, err
	}

	return PbToLeaderboardDefine(resp.Leaderboard), nil
}

func (l *LeaderboardRpc) GetLeaderboardByName(leaderboardName string) (*data.LeaderboardDefine, error) {
	c, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("GetLeaderboardById", "err", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.GetLeaderboard(ctx, &v1.GetLeaderboardRequest{
		LeaderboardId:   "",
		LeaderboardName: leaderboardName,
	})
	if err != nil {
		return nil, err
	}

	return PbToLeaderboardDefine(resp.Leaderboard), nil
}

func (l *LeaderboardRpc) UserScoreChanged(userId string, scoreDelta int64, scoreType string) (map[string]*data.LeaderboardRankData, error) {
	c, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("UserScoreChanged", "err", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.UserScoreChanged(ctx, &v1.UserScoreChangedRequest{
		UserId:     userId,
		ScoreDelta: scoreDelta,
		ScoreType:  scoreType,
	})
	if err != nil {
		return nil, err
	}

	rankData := make(map[string]*data.LeaderboardRankData)
	for _, data := range resp.RankData {
		rank, id := PbToLeaderboardRankData(data)
		rankData[id] = rank
	}

	return rankData, nil
}

func (l *LeaderboardRpc) UserScoreUpdated(userId string, score int64, scoreType string) (map[string]*data.LeaderboardRankData, error) {
	c, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("UserScoreUpdated", "err", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.UserScoreUpdated(ctx, &v1.UserScoreUpdatedRequest{
		UserId:    userId,
		Score:     score,
		ScoreType: scoreType,
	})
	if err != nil {
		return nil, err
	}

	rankData := make(map[string]*data.LeaderboardRankData)
	for _, data := range resp.RankData {
		rank, id := PbToLeaderboardRankData(data)
		rankData[id] = rank
	}

	return rankData, nil
}

func (l *LeaderboardRpc) GetLeaderboardTopN(userId, leaderboardId string, n int32) ([]*data.LeaderboardRankData, error) {
	c, err := Get().GetLeaderboardServiceClient()
	if err != nil {
		zap.S().Errorw("GetLeaderboardTopN", "err", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.GetLeaderboardTopN(ctx, &v1.GetLeaderboardTopNRequest{
		UserId:        userId,
		LeaderboardId: leaderboardId,
		N:             n,
	})
	if err != nil {
		return nil, err
	}

	rankData := make([]*data.LeaderboardRankData, 0)
	for _, data := range resp.RankData {
		rank, _ := PbToLeaderboardRankData(data)
		rankData = append(rankData, rank)
	}

	return rankData, nil
}
