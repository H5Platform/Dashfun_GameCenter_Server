package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/utils"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary	telegram用户登录
// @Tags		User API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=[]data.DashFunUser}	"DashFunUser"
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

	auth, err := utils.CheckAuthorize(c)
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
// @Param		game_id	query	string	true	"游戏Id"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=string}	"DashFunUserId"
// @Router		/api/v1/user/enter_game [get]
func apiEnterGame(c *gin.Context) {
	auth, err := utils.CheckAuthorize(c)
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

type WalletBindReq struct {
	Chain   string `json:"chain" form:"chain" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

// @Summary	用户绑定钱包地址
// @Tags		User API
// @Produce	json
// @Accept	json
// @Param		chain	body	string	true	"网络名称"
// @Param		address	body	string	true	"钱包地址"
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=string}	"DashFunUserId"
// @Router		/api/v1/user/bind_wallet [post]
func apiUserBindWallet(c *gin.Context, user *data.DashFunUser) {
	req := &WalletBindReq{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	u, err := usercenter.Get().UserBindWallet(user.Id, req.Chain, req.Address)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(u.WalletAddress))
}

// @Summary	用户获取最近游戏列表
// @Tags		User API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=[]*data.PlayGameRecord}	"PlayGameRecords"
// @Router		/api/v1/user/play_record [post]
func apiUserGetPlayRecord(c *gin.Context, user *data.DashFunUser) {
	records := usercenter.Get().UserGetPlayRecord(user.Id)
	c.JSON(http.StatusOK, RSuccess(records))
}

// @Summary	用户获取头像数据
// @Tags		User API
// @Produce	json
// @Authorize	"tma {token}"
// @Success	200	{object}	api.JSONResult{data=[]byte}	"avatar png"
// @Router		/api/v1/user/avatar [get]
func apiUserGetHeadPhoto(c *gin.Context, user *data.DashFunUser) {
	headerData := usercenter.Get().GetUserHeadAvatar(user.Id)
	c.Data(http.StatusOK, "image/png", headerData)
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleUser, web.GET, "tg_login", apiTgUserLogin)
	web.GetService().RegisterApi(web.ApiModuleUser, web.GET, "enter_game", apiEnterGame)
	web.GetService().RegisterApi(web.ApiModuleUser, web.POST, "bind_wallet", userHandlerAuthWrapper(apiUserBindWallet))
	web.GetService().RegisterApi(web.ApiModuleUser, web.GET, "play_record", userHandlerAuthWrapper(apiUserGetPlayRecord))
	web.GetService().RegisterApi(web.ApiModuleUser, web.GET, "avatar", userHandlerAuthWrapper(apiUserGetHeadPhoto))
}
