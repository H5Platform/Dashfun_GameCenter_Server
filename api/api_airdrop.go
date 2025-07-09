package api

import (
	"dashfun_gamecenter/AirdropCenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type VestingRequestData struct {
	Existed bool                                `json:"existed"`
	Request *AirdropCenter.CreateVestingRequest `json:"request"`
}

func apiGetAirdropUserDetail(c *gin.Context, user *data.DashFunUser) {
	detail, err := AirdropCenter.Get().GetAirdropUserDetail(user.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(detail))
}

func apiUserClaimAirdrop(c *gin.Context, user *data.DashFunUser) {
	detail, err := AirdropCenter.Get().GetAirdropUserDetail(user.Id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	if detail.Claimed {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("You have already claimed your airdrop."))
		return
	}

	address, existed := c.GetPostForm("address") // 获取用户提交的钱包地址
	if !existed {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("address is required"))
		return
	}

	kcUid, _ := c.GetPostForm("kc_uid")

	if kcUid != "" {
		//填写了kcuid的情况下，只有在tge之前可以提交
		if config.GetConfig().AirdropCfg.GetStartTime() > 0 && time.Now().Unix() > config.GetConfig().AirdropCfg.GetStartTime() {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError("you can only submit your KuCoin ID before the TGE."))
			return
		}
	} else {
		if config.GetConfig().AirdropCfg.GetStartTime() > 0 && time.Now().Unix() < config.GetConfig().AirdropCfg.GetStartTime() {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError("Airdrop has not started yet."))
			return
		}
		if config.GetConfig().AirdropCfg.GetStartTime() > 0 && time.Now().Unix() < (config.GetConfig().AirdropCfg.GetStartTime()+int64(config.GetConfig().AirdropCfg.ClaimTime)) {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError("Airdrop has not started yet."))
			return
		}
	}

	tx, err := AirdropCenter.Get().CreateVestingForUser(user.Id, address, kcUid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(tx))
}

func apiGetUserVestingRequest(c *gin.Context, user *data.DashFunUser) {
	req, ok := AirdropCenter.Get().GetUserCreateVestingRequest(user.Id)
	d := &VestingRequestData{
		Existed: ok,
		Request: req,
	}
	c.JSON(http.StatusOK, RSuccess(d))
}

func apiGetAllKuCoinUsersDetail(c *gin.Context) {
	detail, err := AirdropCenter.Get().GetAllKuCoinUsersDetail()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(detail))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAirdrop, web.GET, "detail", userHandlerAuthWrapper(apiGetAirdropUserDetail))
	web.GetService().RegisterApi(web.ApiModuleAirdrop, web.POST, "claim", userHandlerAuthWrapper(apiUserClaimAirdrop))
	web.GetService().RegisterApi(web.ApiModuleAirdrop, web.GET, "my_request", userHandlerAuthWrapper(apiGetUserVestingRequest))
	web.GetService().RegisterApi(web.ApiModuleAirdrop, web.GET, "kc_users", apiGetAllKuCoinUsersDetail)
}
