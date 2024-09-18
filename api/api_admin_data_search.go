/*
Package api_admin_search
这个包下的api用于后端数据查询
api调用时必须携带Authorization header，值为adminConfig中的backend_password
*/
package api

import (
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type GameResult struct {
	Game   *data.DashFunGame `json:"game"`
	Secret string            `json:"secret"`
}

type UserResult struct {
	User     *data.DashFunUser
	TaskInfo *data.UserTaskInfo
}

func checkAuthorize(c *gin.Context) bool {
	authData := c.GetHeader("authorization")
	if len(authData) == 0 || !strings.HasPrefix(authData, "Bearer ") {
		return false
	}
	s := strings.Split(authData, "Bearer ")[1]
	if s == config.GetConfig().AdminCfg.BackendPassword {
		return true
	}
	return false
}

// apiAdminGetGameInfo
// @Router		/api/v1/admin_search/game/{id} [get]
func apiAdminGetGameInfo(c *gin.Context) {
	if !checkAuthorize(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError("unauthorized"))
		return
	}
	id := c.Param("id")
	game, err := gamecenter.Get().FindGame(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(err.Error()))
		return
	}
	if game == nil {
		//尝试使用name查询
		game, err = gamecenter.Get().FindGameByName(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, RError(err.Error()))
			return
		}
	}

	if game == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(fmt.Sprintf("game %s not found", id)))
		return
	}
	c.JSON(http.StatusOK, RSuccess(&GameResult{
		Game:   game,
		Secret: game.ApiSecret,
	}))
}

// apiAdminGetGameInfo
// @Router		/api/v1/admin_search/user/{id} [get]
func apiAdminGetUserInfo(c *gin.Context) {
	if !checkAuthorize(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError("unauthorized"))
		return
	}

	id := c.Param("id")
	user, err := usercenter.Get().GetDashFunUser(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(err.Error()))
		return
	}
	if user == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(fmt.Sprintf("user %s not found", id)))
		return
	}

	userTaskInfo := taskcenter.Get().GetUserTaskInfo(user, "all")
	c.JSON(http.StatusOK, RSuccess(&UserResult{
		User:     user,
		TaskInfo: userTaskInfo,
	}))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "game/:id", apiAdminGetGameInfo)
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "user/:id", apiAdminGetUserInfo)
}
