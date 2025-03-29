package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/spinwheelcenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserSpinWheelDataResult struct {
	UserId      string                   `json:"user_id"`
	GameId      string                   `json:"game_id"`
	SpinwheelId string                   `json:"spinwheel_id"`
	Rewards     []data.SpinWheelReward   `json:"rewards"`
	RewardIndex int                      `json:"reward_index"` //中奖索引
	Status      data.SpinWheelUserStatus `json:"status"`       //当前状态
}

// @Summary	用户请求转轮盘
// @Tags		SpinWheel API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.SpinWheelUserData}	"用户轮盘数据"
// @Router		/api/v1/spinwheel/spin [get]
func apiUserSpinWheel(c *gin.Context, user *data.DashFunUser) {
	gameId, exist := c.GetQuery("game_id")
	if !exist || gameId == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("param game_id is missing"))
		return
	}
	result, err := spinwheelcenter.Get().UserSpinWheel(user.Id, gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(result))
}

// @Summary	获取用户的轮盘数据
// @Tags		SpinWheel API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=UserSpinWheelDataResult}	"当前轮盘数据及用户状态"
// @Router		/api/v1/spinwheel/get [get]
func apiUserGetSpinWheelInfo(c *gin.Context, user *data.DashFunUser) {
	gameId, exist := c.GetQuery("game_id")
	if !exist || gameId == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("param game_id is missing"))
		return
	}
	center := spinwheelcenter.Get()
	gameWheel, err := center.GetSpinWheelForGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	if gameWheel == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("spinWheel is not existed"))
		return
	}
	userData, err := center.GetSpinWheelUserData(user.Id, gameWheel.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(&UserSpinWheelDataResult{
		UserId:      user.Id,
		GameId:      gameId,
		SpinwheelId: gameWheel.Id,
		Rewards:     gameWheel.Rewards,
		RewardIndex: userData.RewardIndex,
		Status:      userData.Status,
	}))
}

// @Summary	用戶领取转盘奖励
// @Tags		SpinWheel API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.SpinWheelReward}	"当前轮盘数据及用户状态"
// @Router		/api/v1/spinwheel/claim [get]
func apiUserClaimSpinWheelReward(c *gin.Context, user *data.DashFunUser) {
	gameId, exist := c.GetQuery("game_id")
	if !exist || gameId == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("param game_id is missing"))
		return
	}
	center := spinwheelcenter.Get()
	reward, err := center.UserClaimReward(user.Id, gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(reward))
}
func init() {
	web.GetService().RegisterApi(web.ApiModuleSpinWheel, web.GET, "get", userHandlerAuthWrapper(apiUserGetSpinWheelInfo))
	web.GetService().RegisterApi(web.ApiModuleSpinWheel, web.GET, "spin", userHandlerAuthWrapper(apiUserSpinWheel))
	web.GetService().RegisterApi(web.ApiModuleSpinWheel, web.GET, "claim", userHandlerAuthWrapper(apiUserClaimSpinWheelReward))
}
