package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserGameSearchRequest struct {
	Keyword string `form:"keyword"`
	Genre   []int  `json:"genre" form:"genre"` //游戏类型Id
	Size    int64  `form:"size"`
	Page    int64  `form:"page"`
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

// @Summary	获取游戏类型数据
// @Tags		Games API
// @Produce	json
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=[]data.DashFunGameGenre}	"DashFunGameGenre"
// @Router		/api/v1/game/genres [get]
func apiUserGetGenres(c *gin.Context) {
	c.JSON(http.StatusOK, RSuccess(gamecenter.Get().GetGameGenres()))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, ":id", apiUserStartGame)
	web.GetService().RegisterApi(web.ApiModuleGame, web.POST, "search", userHandlerAuthWrapper(apiUserFindGames))
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, "genres", apiUserGetGenres)
}
