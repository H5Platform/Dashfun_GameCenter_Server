package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type UserGameSearchRequest struct {
	Keyword string `form:"keyword"`
	Genre   []int  `json:"genre" form:"genre"` //游戏类型Id
	Size    int64  `form:"size"`
	Page    int64  `form:"page"`
}

type UserGameDataRequest struct {
	Key  string `form:"key" binding:"required"`
	Data string `form:"data"`
}

type UserGetDataResult struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

// @Summary	telegram用户开启游戏
// @Tags		Games API
// @Produce	json
// @Param	id path string true "开启的游戏Id"
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=[]data.DashFunGame}	"DashFunGame"
// @Router		/api/v1/game/{id} [get]
func apiUserStartGame(c *gin.Context) {
	id := c.Param("id")
	game, err := gamecenter.Get().FindGame(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(err.Error()))
		return
	}
	if game == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(fmt.Sprintf("game %s not found", id)))
		return
	}
	c.JSON(http.StatusOK, RSuccess(game))
}

// @Summary	用户搜索游戏
// @Tags		Games API
// @Produce	json
// @Param		keyword	body		string									false	"查询关键字"
// @Param		genre	body		[]int									false	"查询类型"
// @Param		size	body		int64									false	"每页数量"
// @Param		page	body		int64									false	"当前页数，从1开始"
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=[]data.DashFunGame}	"DashFunGame"
// @Router		/api/v1/game/search [post]
func apiUserFindGames(c *gin.Context, user *data.DashFunUser) {
	req := &UserGameSearchRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	games, totalPages, err := gamecenter.Get().FindGames(req.Keyword, req.Genre, req.Size, req.Page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, PageSuccess(games, req.Page, req.Size, totalPages))
}

// @Summary	获取所有测试游戏
func apiGetTestingGames(c *gin.Context, user *data.DashFunUser) {
	games, totalPages, err := gamecenter.Get().FindGamesBackend("", nil, data.DashFunGameStatus_Pending, 1000, 1)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, PageSuccess(games, 1, 100, totalPages))
}

// @Summary	获取游戏类型数据
// @Tags		Games API
// @Produce	json
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=[]data.DashFunGameGenre}	"DashFunGameGenre"
// @Router		/api/v1/game/genres [get]
func apiUserGetGenres(c *gin.Context) {
	c.JSON(http.StatusOK, RSuccess(gamecenter.Get().GetGameGenres()))
}

// @Summary 用户保存数据
// @Tags		Games API
// @Produce	json
// @Authorize "tma {token}"
// @Param		key		body		string								true	"数据存储键值"
// @Param		data	body		string								true	"要存储的数据"
// @Success	200		{object}	api.JSONResult{data=bool}	"save result"
// @Router		/api/v1/game/{id}/data [post]
func apiUserSetData(c *gin.Context, user *data.DashFunUser) {
	gameId := c.Param("id")
	req := &UserGameDataRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	game, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	if game == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(fmt.Sprintf("game %s not found", gameId)))
		return
	}
	_, err = usercenter.Get().UserSaveData(user.Id, gameId, req.Key, req.Data, game.IsTesting())
	zap.S().Infow("save user data", "user", user.Id, "key", req.Key, "save", req.Data)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(req.Key))
}

// @Summary 用户读取数据
// @Tags		Games API
// @Produce	json
// @Authorize "tma {token}"
// @Param		key		query		string								true	"要读取的数据键值"
// @Success	200		{object}	api.JSONResult{data=string}	"save data"
// @Router		/api/v1/game/{id}/data [get]
func apiUserGetData(c *gin.Context, user *data.DashFunUser) {
	gameId := c.Param("id")
	key, exist := c.GetQuery("key")
	if !exist {
		c.AbortWithStatusJSON(http.StatusNotFound, RError("param key not found"))
		return
	}
	game, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	if game == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(fmt.Sprintf("game %s not found", gameId)))
		return
	}
	saveData, err := usercenter.Get().UserGetData(user.Id, gameId, key, game.IsTesting())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(saveData))
}

// @Summary 用户读取数据，同时返回key和data
// @Tags		Games API
// @Produce	json
// @Authorize "tma {token}"
// @Param		key		query		string								true	"要读取的数据键值"
// @Success	200		{object}	api.JSONResult{data=UserGetDataResult}	"save data"
// @Router		/api/v1/game/{id}/data_v2 [get]
func apiUserGetData1(c *gin.Context, user *data.DashFunUser) {
	gameId := c.Param("id")
	key, exist := c.GetQuery("key")
	if !exist {
		c.AbortWithStatusJSON(http.StatusNotFound, RError("param key not found"))
		return
	}
	game, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	if game == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(fmt.Sprintf("game %s not found", gameId)))
		return
	}
	saveData, err := usercenter.Get().UserGetData(user.Id, gameId, key, game.IsTesting())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(&UserGetDataResult{Key: key, Data: saveData}))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, ":id", apiUserStartGame)
	web.GetService().RegisterApi(web.ApiModuleGame, web.POST, "search", userHandlerAuthWrapper(apiUserFindGames))
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, "genres", apiUserGetGenres)
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, "testing", userHandlerAuthWrapper(apiGetTestingGames))

	web.GetService().RegisterApi(web.ApiModuleGame, web.POST, ":id/data", userHandlerAuthWrapper(apiUserSetData))
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, ":id/data", userHandlerAuthWrapper(apiUserGetData))
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, ":id/data_v2", userHandlerAuthWrapper(apiUserGetData1))
}
