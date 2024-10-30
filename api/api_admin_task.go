package api

import (
	"dashfun_gamecenter/admin"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type adminUpdateTaskRequest struct {
	Id        string                    `json:"id" form:"id" required:"false"`
	Name      string                    `json:"name" form:"name" required:"true"`
	Type      data.DashFunTaskType      `json:"type" form:"type" required:"true"`
	GameId    string                    `json:"game_id" form:"game_id" required:"false"`
	Category  data.DashFunTaskCategory  `json:"category" form:"category" required:"true"`
	Condition data.DashFunTaskCondition `json:"condition" form:"condition" required:"true"`
	Reward    data.DashFunTaskReward    `json:"reward" form:"reward" required:"true"`
	Open      bool                      `json:"open" form:"open" required:"false"`
}

// apiAdminGameSearch
//
//	@Summary	获取指定游戏绑定的任务
//	@Tags		Admin API
//	@Produce	json
//	@Param		game_id	path		string										true	"游戏id"
//	@Success	200		{object}	api.JSONResult{data=[]data.DashFunTaskData}	"Search Result"
//	@Router		/api/v1/admin/task/get/{game_id} [post]
func apiAdminTaskGetForGame(c *gin.Context, op *admin.AdminUser) {
	gameId := c.Param("game_id")
	if gameId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param game_id is required"))
		return
	}
	tasks := taskcenter.Get().GetGameTasksBackend(gameId)
	c.JSON(http.StatusOK, RSuccess(tasks))
}

// @Summary	搜索游戏
// @Tags		Admin API
// @Accept		json
// @Produce	json
// @Param		id			body		string										true	"更新的任务Id"
// @Param		name		body		string										true	"任务名称"
// @Param		type		body		data.DashFunTaskType						true	"任务类型"
// @Param		category	body		data.DashFunTaskCategory					true	"任务分类"
// @Param		condition	body		data.DashFunTaskCondition					true	"任务分类"
// @Param		reward		body		data.DashFunTaskReward						true	"任务奖励"
// @Param		open		body		bool										true	"任务是否开放"
// @Success	200			{object}	api.JSONResult{data=[]data.DashFunTaskData}	"更新后的任务数据"
// @Router		/api/v1/admin/task/update [post]
func apiAdminTaskUpdate(c *gin.Context, op *admin.AdminUser) {
	req := &adminUpdateTaskRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	task, err := taskcenter.Get().UpdateTask(req.Id, req.Name, req.Type, req.Category, req.Condition, req.Reward, req.Open)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(task))
}

// @Summary	搜索游戏
// @Tags		Admin API
// @Accept		json
// @Produce	json
// @Param		name		body		string										true	"任务名称"
// @Param		game_id		body		string										true	"绑定的游戏Id"
// @Param		type		body		data.DashFunTaskType						true	"任务类型"
// @Param		category	body		data.DashFunTaskCategory					true	"任务分类"
// @Param		condition	body		data.DashFunTaskCondition					true	"任务分类"
// @Param		reward		body		data.DashFunTaskReward						true	"任务奖励"
// @Success	200			{object}	api.JSONResult{data=[]data.DashFunTaskData}	"新增的任务数据"
// @Router		/api/v1/admin/task/create [post]
func apiAdminTaskCreate(c *gin.Context, op *admin.AdminUser) {
	req := &adminUpdateTaskRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	task, err := taskcenter.Get().CreateTaskAutoId(req.Name, req.GameId, req.Type, req.Category, req.Condition, req.Reward)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(task))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "task/create", adminHandlerAuthWrapper(admin.AdminAuth_Task, apiAdminTaskCreate))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "task/update", adminHandlerAuthWrapper(admin.AdminAuth_Task, apiAdminTaskUpdate))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "task/get/:game_id", adminHandlerAuthWrapper(admin.AdminAuth_Task, apiAdminTaskGetForGame))
}
