/*
Package api_admin_search
这个包下的api用于后端数据查询
api调用时必须携带Authorization header，值为adminConfig中的backend_password
*/
package api

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/invitecenter"
	"dashfun_gamecenter/taskcenter"
	"dashfun_gamecenter/tgbot"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"net/http"
	"strings"
	"time"
)

type GameResult struct {
	Game   *data.DashFunGame `json:"game"`
	Secret string            `json:"secret"`
}

type UserResult struct {
	User     *data.DashFunUser
	TaskInfo *data.UserTaskInfo
}

type CreateDFUserRequest struct {
	From     data.DashFunUserFrom `json:"from" form:"from" binding:"required"`
	Username string               `json:"username" form:"username" binding:"required"`
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
//
//	@Router	/api/v1/admin_search/game/{id} [get]
func apiAdminGetGameInfo(c *gin.Context) {
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
//
//	@Router	/api/v1/admin_search/user/{id} [get]
func apiAdminGetUserInfo(c *gin.Context) {
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

func apiAdminGetUserCoins(c *gin.Context) {
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

	coins := coincenter.Get().GetDashFunCoins()

	userCoins := make(map[string]int32)
	for _, coin := range coins {
		ud := coincenter.Get().GetCoinUserData(user.Id, coin.Id)
		if ud != nil {
			userCoins[coin.Name] = ud.Amount
		}
	}

	c.JSON(http.StatusOK, RSuccess(userCoins))
}

func apiAdminSearchUserCoinRecords(c *gin.Context) {
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

	coin := c.Query("coin_name")
	if len(coin) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("coin_name is required"))
		return
	}

	coinData, exist := coincenter.Get().GetCoinByName(coin)
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(fmt.Sprintf("coin %s not found", coin)))
		return
	}

	ucd := coincenter.Get().GetCoinUserData(user.Id, coinData.Id)
	records := coincenter.Get().GetUserCoinRecords(id, coinData.Id, 0)

	str := fmt.Sprintf("## User **[%s]%s** Coin **[%s]** Records\n\n", user.Id, user.DisplayName, coinData.Name)
	str += fmt.Sprintf("|%s|%s|%s|%s|%s|\n", "Time", "Type", "Change", "Reason", "Info")
	str += fmt.Sprintf("|%s|%s|%s|%s|%s|\n", "---", "---", "---", "---", "---")
	var totalAdd int32
	var totalDec int32
	var totalBalance int32
	for _, record := range records {
		t := time.UnixMilli(record.Time).Format("2006-01-02 15:04:05")
		recordType := "Add"
		if record.Reason == coincenter.CoinAddReasonRecalculate {
			recordType = "Recalculate"
			str += fmt.Sprintf("|%s|%s|%d|%s|%s|\n", t, recordType, record.Change, record.Reason, record.Info)
			continue
		} else if record.Change > 0 {
			totalAdd += record.Change
		} else if record.Change < 0 {
			recordType = "Dec"
			totalDec += -record.Change
		}
		totalBalance += record.Change
		str += fmt.Sprintf("|%s|%s|%d|%s|%s|\n", t, recordType, record.Change, record.Reason, record.Info)
	}

	str += fmt.Sprintf("- User **[%s]%s** Coin **[%s]** Total Add: %d Total Dec:%d Record Balance:%d\n", user.Id, user.DisplayName, coinData.Name, totalAdd, totalDec, totalBalance)
	str += fmt.Sprintf("- User **[%s]%s** Coin **[%s]** Current Balance: %d\n", user.Id, user.DisplayName, coinData.Name, ucd.Amount)

	if totalBalance > ucd.Amount {
		str += fmt.Sprintf("\n\n## <font color=\"red\">**Warning**</font> User **[%s]%s** Coin **[%s]** Record Balance: %d > Current Balance: %d\n", user.Id, user.DisplayName, coinData.Name, totalBalance, ucd.Amount)
	}

	html := markdown.ToHTML([]byte(str), nil, nil)
	c.Header("Content-Disposition", "inline; filename=coin_records.html")
	c.Data(200, "text/html; charset=utf-8", html)
}

// apiAdminRecalculateUserCoin 根据用户的coin记录重新计算用户的coin余额
func apiAdminRecalculateUserCoin(c *gin.Context) {
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

	coin := c.Query("coin_name")
	if len(coin) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("coin_name is required"))
		return
	}

	coinData, exist := coincenter.Get().GetCoinByName(coin)
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(fmt.Sprintf("coin %s not found", coin)))
		return
	}

	ucd := coincenter.Get().GetCoinUserData(user.Id, coinData.Id)

	_, _, totalBalance := coincenter.Get().CalculateUserCoinByRecords(user.Id, coinData.Id)

	before := ucd.Amount
	makeup := totalBalance - ucd.Amount
	if totalBalance > ucd.Amount {
		coincenter.Get().AddUserCoinAmount(user.Id, coinData.Id, makeup, coincenter.CoinAddReasonRecalculate, fmt.Sprintf("Make Up %d Coin", makeup))
		c.JSON(http.StatusOK, gin.H{
			"coin":    gin.H{"id": coinData.Id, "name": coinData.Name},
			"user":    gin.H{"id": user.Id, "name": user.DisplayName},
			"balance": gin.H{"before": before, "make up": makeup, "current": totalBalance},
		})
	} else {
		c.JSON(http.StatusOK, "no make up needed")
	}
}

func apiAdminCreateDashFunKolUser(c *gin.Context) {

	username := c.PostForm("username")

	if len(username) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("username is required"))
		return
	}

	user, err := usercenter.Get().CreateDashFunUser(data.DF_UserFrom_Kol, username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(user))
}

func apiAdminGetKolUsers(c *gin.Context) {
	users, err := usercenter.Get().GetUsersFrom(data.DF_UserFrom_Kol)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, RError(err.Error()))
		return
	}

	str := `
<style>
	table {
		width: auto;
		min-width:600px;
	}
	th, td {
		padding: 8px;
	}
	tr:nth-child(even) {
		background-color: #f2f2f2;
	}
</style>
`

	str += fmt.Sprintf("## Kol Invite Records\n\n")

	str += fmt.Sprintf("- **Invited(Total)**: %s\n", "Total Invited")
	str += fmt.Sprintf("- **Actived**: %s\n", "Users Invited who have reached 5000 xp (Including new users)")
	str += fmt.Sprintf("- **New Users**: %s\n\n", "New users invited by this KOL who have reached 5000 xp")

	str += fmt.Sprintf("|%s|%s|%s|%s|%s|%s\n", "UserId", "Name", "Invited(Total)", "Actived", "New Users", "Link")
	str += fmt.Sprintf("|%s|%s|%s|%s|%s|%s\n", "---", "---", ":---:", ":---:", ":---:", "---")

	for _, user := range users {
		invited, _ := invitecenter.Get().GetInvitedUsers(user.Id)

		invitedCount := 0
		activatedCount := 0
		newUserCount := 0
		if invited != nil {
			invitedCount = len(invited)
		}

		for _, userData := range invited {
			if userData.InvitedStatus == data.InvitedStatus_Success {
				activatedCount++
				if userData.InvitedType == data.InvitedType_NewUser {
					newUserCount++
				}
			}
		}

		inviteLink := tgbot.InviteLink(user.Id)

		str += fmt.Sprintf("|%s|%s|**%d**|**%d**|**%d**|%s\n", user.Id, user.DisplayName, invitedCount, activatedCount, newUserCount, inviteLink)
	}

	html := markdown.ToHTML([]byte(str), nil, nil)
	c.Header("Content-Disposition", "inline; filename=coin_records.html")
	c.Data(200, "text/html; charset=utf-8", html)
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "game/:id", backendAdminAuthWrapper(apiAdminGetGameInfo))
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "user/:id", backendAdminAuthWrapper(apiAdminGetUserInfo))
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "user/:id/coins", backendAdminAuthWrapper(apiAdminGetUserCoins))
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "user/:id/coin/record", backendAdminAuthWrapper(apiAdminSearchUserCoinRecords))
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "user/:id/coin/recalculate", backendAdminAuthWrapper(apiAdminRecalculateUserCoin))
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.POST, "user/create/kol", backendAdminAuthWrapper(apiAdminCreateDashFunKolUser))
	web.GetService().RegisterApi(web.ApiModuleAdminSearch, web.GET, "user/invited/kol", apiAdminGetKolUsers)

}

func backendAdminAuthWrapper(handler func(ctx *gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		if !checkAuthorize(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError("unauthorized"))
			return
		}

		if handler != nil {
			handler(c)
		}
	}
}
