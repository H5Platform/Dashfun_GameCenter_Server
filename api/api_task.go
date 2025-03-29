package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary	用户请求任务奖励
// @Tags		Task API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Param		task_id	query	string	true	"任务Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.DashFunTaskUserData}	"DashFunTaskUserData"
// @Router		/api/v1/task/claim [get]
func apiUserClaimTaskReward(c *gin.Context, user *data.DashFunUser) {
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

	_ = gameId

	//auth, err := CheckAuthorize(c)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	//
	//user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}

	r, err := taskcenter.Get().UserClaimReward(user, taskId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(r))
}

// @Summary	获取用户各个状态的任务数量
// @Tags		Task API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=map[int]int}	"UserTaskCount"
// @Router		/api/v1/task/count [get]
func apiGetUserTaskCount(c *gin.Context, user *data.DashFunUser) {
	gameId, exist := c.GetQuery("game_id")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param game_id is required"))
		return
	}
	//auth, err := CheckAuthorize(c)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	//
	//user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	info := taskcenter.Get().GetUserTaskInfo(user, gameId)

	//各个状态下的任务数量
	r := make(map[int]int)

	for _, ud := range info.UserData {
		switch ud.Status {
		case data.TaskStatus_Verify_Pending, data.TaskStatus_InProgress:
			r[int(data.TaskStatus_InProgress)] += 1
			break
		case data.TaskStatus_Completed:
			r[int(data.TaskStatus_Completed)] += 1
			break
		case data.TaskStatus_Claimed:
			r[int(data.TaskStatus_Claimed)] += 1
			break
		}
	}

	c.JSON(http.StatusOK, RSuccess(r))
}

// @Summary	获取用户的任务信息
// @Tags		Task API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.UserTaskInfo}	"UserTaskInfo"
// @Router		/api/v1/task/list [get]
func apiGetUserTaskInfo(c *gin.Context, user *data.DashFunUser) {
	gameId, exist := c.GetQuery("game_id")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param game_id is required"))
		return
	}
	//auth, err := CheckAuthorize(c)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	//
	//user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	info := taskcenter.Get().GetUserTaskInfo(user, gameId)
	c.JSON(http.StatusOK, RSuccess(info))
}

// @Summary	任务条目被点击
// @Tags		Task API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Param		task_id	query	string	true	"任务Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.DashFunTaskUserData}	"UserTaskInfo"
// @Router		/api/v1/task/clicked [get]
func apiOnTaskClicked(c *gin.Context, user *data.DashFunUser) {
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

	//auth, err := CheckAuthorize(c)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	//
	//user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}

	//on clicked
	userData, err := taskcenter.Get().UserClickedTask(user, taskId, gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(userData))
}

// @Summary	加入tg任务验证
// @Tags		Task API
// @Produce	json
// @Param		game_id	query	string	true	"游戏Id"
// @Param		task_id	query	string	true	"任务Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=data.DashFunTaskUserData}	"UserTaskInfo"
// @Router		/api/v1/task/verify [get]
func apiVerifyUserTask(c *gin.Context, user *data.DashFunUser) {
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

	//auth, err := CheckAuthorize(c)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}
	//
	//user, err := usercenter.Get().GetDashFunUserByTgAuthData(auth.Token)
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
	//	return
	//}

	//verify task
	userData, err := taskcenter.Get().UserVerifyTask(user, taskId, gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(userData))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "list", userHandlerAuthWrapper(apiGetUserTaskInfo))
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "verify", userHandlerAuthWrapper(apiVerifyUserTask))
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "clicked", userHandlerAuthWrapper(apiOnTaskClicked))
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "claim", userHandlerAuthWrapper(apiUserClaimTaskReward))
	web.GetService().RegisterApi(web.ApiModuleTask, web.GET, "count", userHandlerAuthWrapper(apiGetUserTaskCount))
}
