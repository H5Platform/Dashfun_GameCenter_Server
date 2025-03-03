package api

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/invitecenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type InvitedUsersResult struct {
	RewardPoint []config.RewardPoint    `json:"reward_point"` //每个用户奖励的分数
	Friends     []*data.InvitedUserData `json:"friends"`      //邀请用户信息，只显示50个最多
}

func apiGetMyFriends(c *gin.Context, user *data.DashFunUser) {
	friends, err := invitecenter.Get().GetInvitedUsers(user.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	//只发前50个
	if len(friends) > 50 {
		friends = friends[:50]
	}

	rewardPoints := config.GetConfig().InviteCfg.PointReward

	c.JSON(http.StatusOK, RSuccess(&InvitedUsersResult{
		RewardPoint: rewardPoints,
		Friends:     friends,
	}))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleFriends, web.GET, "/my", userHandlerAuthWrapper(apiGetMyFriends))
}
