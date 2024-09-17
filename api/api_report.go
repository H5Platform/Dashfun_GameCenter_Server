package api

import (
	"crypto/md5"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

// server report相关api
// 由游戏服务器发来的用户相关数据，例如用户升级，用户装备提升到几星，等等，按需增加

// verifyGame
// 加密方式，md5(query string + &timestamp=unix_mill + & + secretKey)
// 参数按照字母升序排列
func verifyGame(params map[string]string, game *data.DashFunGame, md5Value string) bool {
	params["secret"] = game.ApiSecret

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var queryString = ""

	for _, k := range keys {
		if queryString != "" {
			queryString += "&"
		}
		queryString += k + "=" + url.QueryEscape(params[k])
	}

	md5v := md5.Sum([]byte(queryString))
	md5Str := fmt.Sprintf("%x", md5v)

	return md5Str == md5Value
}

type GameReportReq struct {
	GameId    string `json:"game_id" form:"game_id" binding:"required"`
	UserId    string `json:"user_id" form:"user_id" binding:"required"`
	Level     int    `json:"level" form:"level" binding:"required"`
	Timestamp string `json:"timestamp" form:"timestamp" binding:"required"`
	Sign      string `json:"sign" form:"sign" binding:"required"`
}

// @Summary	游戏方上报玩家等级，Server端使用
// @Tags		Game Report API
// @Produce	json
// @Param		game_id		query		string						true	"上报的游戏Id"
// @Param		user_id		query		string						true	"上报的用户Id"
// @Param		level		query		int							true	"用户等级"
// @Param		timestamp	query		string						true	"时间戳"
// @Param		sign		query		string						true	"签名字符串"
// @Success	200			{object}	api.JSONResult{data=string}	"返回结果"
// @Router		/api/v1/game_report/player_level [get]
func apiReportPlayerLevel(c *gin.Context) {

	req := &GameReportReq{}
	err := c.ShouldBindQuery(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	game, err := gamecenter.Get().FindGame(req.GameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	err = nil

	user, err := usercenter.Get().GetDashFunUser(req.UserId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	params := make(map[string]string)
	params["game_id"] = req.GameId
	params["user_id"] = req.UserId
	params["level"] = strconv.Itoa(req.Level)
	params["timestamp"] = req.Timestamp

	b := verifyGame(params, game, req.Sign)

	if b {
		events.PlayerLevelUpEvents.Emit(&events.EventPlayerLevelUp{
			User:  user,
			Game:  game,
			Level: req.Level,
		})
		c.JSON(http.StatusOK, RSuccess(""))
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("Error Signature"))
	}
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleGameReport, web.GET, "player_level", apiReportPlayerLevel)
}

//func init() {
//	p := map[string]string{
//		"gameId":    "123456",
//		"timestamp": "3215123",
//		"wwww":      "dddd",
//	}
//	game, err := dao.GetGameDao().GetGameByName("Stone Age")
//	if err != nil {
//		log.Fatal(err)
//	}
//	v := verifyGame(p, game, "asdf")
//	log.Printf("%v", v)
//}
