package api

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/leaderboardcenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LeaderboardMyInfo struct {
	MyRank  int64 `json:"my_rank"`
	MyPoint int64 `json:"my_point"`
}

// @Summary	获取排行榜的Top100，用户自己的排行会在最后
// @Tags	Leaderboard API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=[]leaderboardcenter.LeaderboardData}	""
// @Router		/api/v1/leaderboard/xp_top [get]
func apiLeaderboardTop20(c *gin.Context, user *data.DashFunUser) {
	top, err := leaderboardcenter.Get().GetTop(20)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	rank, score, err := leaderboardcenter.Get().GetUserRankAndScore(user.Id)
	if err != nil || score < int64(config.GetConfig().LeaderboardBotCfg.RecordScoreMin) {
		// 5000分以下不显示
		top = append(top, &leaderboardcenter.LeaderboardData{
			Id:          user.Id,
			Rank:        0,
			Score:       score,
			UserName:    user.UserName,
			DisplayName: user.DisplayName,
		})
	} else {
		top = append(top, &leaderboardcenter.LeaderboardData{
			Id:          user.Id,
			Rank:        rank,
			Score:       score,
			UserName:    user.UserName,
			DisplayName: user.DisplayName,
			Avatar:      user.AvatarUrl,
		})
	}

	c.JSON(http.StatusOK, RSuccess(top))
}

// @Summary	返回用户在排行榜中的排名信息
// @Tags	Leaderboard API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=api.LeaderboardMyInfo}	"UploadFaceResult"
// @Router		/api/v1/leaderboard/me [get]
func apiLeaderboardMe(c *gin.Context, user *data.DashFunUser) {
	rank, score, err := leaderboardcenter.Get().GetUserRankAndScore(user.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(LeaderboardMyInfo{
		MyRank:  rank,
		MyPoint: score,
	}))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleLeaderboard, web.GET, "/xp_top", userHandlerAuthWrapper(apiLeaderboardTop20))
	web.GetService().RegisterApi(web.ApiModuleLeaderboard, web.GET, "/me", userHandlerAuthWrapper(apiLeaderboardMe))
}
