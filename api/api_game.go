package api

import (
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

func init() {
	web.GetService().RegisterApi(web.ApiModuleGame, web.GET, ":id", apiUserStartGame)
}
