package api

import (
	"dashfun_gamecenter/admin"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/tencentcos"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminCoinDataRequest struct {
	Id          string            `json:"id" form:"id"`
	Name        string            `json:"name" form:"name"`
	Symbol      string            `json:"symbol" form:"symbol"`
	Desc        string            `json:"desc" form:"desc"`
	CanWithdraw bool              `json:"can_withdraw" form:"can_withdraw"`
	MinWithdraw float32           `json:"min_withdraw" form:"min_withdraw"`
	ChainAddr   map[string]string `json:"chain_addr" form:"chain_addr"`
}

// apiAdminCoinGetByGame
//
//	@Summary	获取指定id的游戏数据
//	@Tags		Admin API
//	@Produce	json
//	@Param		game_id	path		string									true	"游戏ID"
//	@Success	200		{object}	api.JSONResult{data=data.CoinData}	"Search Result"
//	@Router		/api/v1/admin/coin/get/{game_id} [get]
func apiAdminCoinGetByGame(c *gin.Context, op *admin.AdminUser) {
	gameId := c.Param("game_id")
	if gameId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("game id is required"))
		return
	}
	_, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	coin, existed := coincenter.Get().GetCoinByGame(gameId)
	if !existed {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError("coin not found"))
		return
	}

	c.JSON(http.StatusOK, RSuccess(coin))
}

// apiAdminCoinUpdate
//
//	@Summary	获取指定id的游戏数据
//	@Tags		Admin API
//	@Produce	json
//	@Param		game_id	path		string									true	"游戏ID"
//	@Success	200		{object}	api.JSONResult{data=data.CoinData}	"Search Result"
//	@Router		/api/v1/admin/coin/update/{game_id} [post]
func apiAdminCoinUpdate(c *gin.Context, op *admin.AdminUser) {
	gameId := c.Param("game_id")
	if gameId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("game id is required"))
		return
	}
	_, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	req := &AdminCoinDataRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	// Check for duplicate coin name
	c1, existed := coincenter.Get().GetCoinByName(req.Name)
	if existed && c1.Id != req.Id {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("coin name already exists"))
		return
	}

	coin, err := coincenter.Get().UpdateCoin(req.Id, req.Name, req.Desc, req.Symbol, req.CanWithdraw, req.MinWithdraw, req.ChainAddr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	//upload coin icon
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	iconImg, ok := form.File["icon"]
	if ok {
		imgFile := iconImg[0]
		src, err := imgFile.Open()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
			return
		}
		defer src.Close()
		p := make([]byte, imgFile.Size)
		_, err = src.Read(p)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
			return
		}

		key := "images/" + coin.BindGameId + "/coin.png"
		_, err = tencentcos.Get().UploadData(key, p, "image/png")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, RSuccess(coin))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.GET, "coin/get/:game_id", adminHandlerAuthWrapper(admin.AdminAuth_Game, apiAdminCoinGetByGame))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "coin/update/:game_id", adminHandlerAuthWrapper(admin.AdminAuth_Game, apiAdminCoinUpdate))
}
