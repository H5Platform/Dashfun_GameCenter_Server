package api

import (
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary 获取用户的任务信息
// @Tags		Task API
// @Produce	json
// @Param	game_id query string true "游戏Id"
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=data.UserTaskInfo}	"UserTaskInfo"
// @Router		/api/v1/task/list [get]
func apiGetUserTaskInfo(c *gin.Context) {
	gameId, exist := c.GetQuery("game_id")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param game_id is required"))
		return
	}
	auth, err := CheckAuthorize(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}
	info := taskcenter.Get().GetUserTaskInfo(user.Id, gameId)
	c.JSON(http.StatusOK, RSuccess(info))
}

// @Summary 加入tg任务验证
// @Tags		Task API
// @Produce	json
// @Param	game_id query string true "游戏Id"
// @Param	task_id query string true "任务Id"
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=data.DashFunTaskUserData}	"UserTaskInfo"
// @Router		/api/v1/task/tg_verify [get]
func apiVerifyUserTGChannelTask(c *gin.Context) {
	taskId, exist := c.GetQuery("task_id")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param task_id is required"))
		return
	}
	gameId, exist := c.GetQuery("game_id")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param game_id is required"))
		return
	}

	auth, err := CheckAuthorize(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	//verify tg task
	userData, err := taskcenter.Get().UserVerifyTGChannel(user, taskId, gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(userData))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "list", apiGetUserTaskInfo)
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "tg_verify", apiVerifyUserTGChannelTask)
}
