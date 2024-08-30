package api

import (
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary	telegram用户登录
// @Tags		Telegram API
// @Produce	json
// @Authorize "tma {token}"
// @Success	200		{object}	api.JSONResult{data=[]data.DashFunUser}	"DashFunUser"
// @Router		/api/v1/user/tg_login [get]
func apiTgUserLogin(c *gin.Context) {
	//idStr, ok := c.GetQuery("id")
	//if !ok {
	//	c.AbortWithStatus(http.StatusBadRequest)
	//	return
	//}
	//
	//id, err := strconv.Atoi(idStr)
	//if err != nil {
	//	c.AbortWithStatus(http.StatusBadRequest)
	//	return
	//}

	auth, err := CheckAuthorize(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	login, err := usercenter.Get().TGUserLogin(auth.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(login.User))
}

// @Summary	用户点击了play
// @Tags		User API
// @Produce	json
// @Param	game_id query string true "游戏Id"
// @Authorize "tma {token}"
// @Success	200		{string} "DashFunUserId"
// @Router		/api/v1/user/enter_game [get]
func apiEnterGame(c *gin.Context) {
	auth, err := CheckAuthorize(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}

	gameId, exist := c.GetQuery("game_id")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("param game_id is required"))
	}

	user, err := usercenter.Get().UserEnterGame(auth.Token, gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(user.Id))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleUser, web.GET, "tg_login", apiTgUserLogin)
	web.GetService().RegisterApi(web.ApiModuleUser, web.GET, "enter_game", apiEnterGame)
}
