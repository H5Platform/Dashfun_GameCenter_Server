package api

import (
	"dashfun_gamecenter/accountcenter"
	"dashfun_gamecenter/nacoscenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AccountCreateRequest struct {
	Username string                         `json:"username" form:"username" binding:"required"`
	Password string                         `json:"password" form:"password" binding:"required"`
	AccType  nacoscenter.DashFunAccountType `json:"acc_type" form:"acc_type" binding:"required"` //账号类型，默认是email
}

type AccountCreateResponse struct {
	AccountId string                           `json:"account_id"` //账号Id
	Username  string                           `json:"username"`
	Type      nacoscenter.DashFunAccountType   `json:"type"`   //账号类型，默认是email
	Status    nacoscenter.DashFunAccountStatus `json:"status"` //账号状态
}

type AccountLoginResponse struct {
	AccountId   string                           `json:"account_id"` //账号Id
	Username    string                           `json:"username"`
	Type        nacoscenter.DashFunAccountType   `json:"type"`   //账号类型，默认是email
	Status      nacoscenter.DashFunAccountStatus `json:"status"` //账号状态
	Token       string                           `json:"token"`  //登录token
	DisplayName string                           `json:"display_name"`
}

type AccountVerifyEmailRequest struct {
	AccountId string `json:"account_id" form:"account_id" binding:"required"`
	Code      string `json:"code" form:"code"  `
}

type CheckTokenRequest struct {
	AccountId string                         `json:"account_id" form:"account_id" binding:"required"`
	Token     string                         `json:"token" form:"token" binding:"required"`
	AccType   nacoscenter.DashFunAccountType `json:"acc_type" form:"acc_type" binding:"required"`
}

type RequestResetPasswordRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
}

type ResetPasswordRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Code     string `json:"code" form:"code" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

func apiAccountCreate(c *gin.Context) {
	req := &AccountCreateRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	account, err := accountcenter.Get().CreateAccount(req.Username, req.Password, req.AccType)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(&AccountCreateResponse{
		AccountId: account.AccountId,
		Username:  account.Username,
		Type:      account.Type,
		Status:    account.Status,
	}))
}

func apiAccountLogin(c *gin.Context) {
	req := &AccountCreateRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	// 账号登录，如果账号处于未验证状态，会返回account，其他错误account=nil
	account, err := accountcenter.Get().LoginAccount(req.Username, req.Password, req.AccType)
	if account == nil && err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	if account != nil {
		c.JSON(http.StatusOK, RSuccess(&AccountLoginResponse{
			AccountId:   account.AccountId,
			Username:    account.Username,
			Type:        account.Type,
			Status:      account.Status,
			DisplayName: account.DisplayName,
			Token:       account.Token,
		}))
	}
}

func apiAccountSendVerifyEmail(c *gin.Context) {
	req := &AccountVerifyEmailRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	err = accountcenter.Get().RequestSendVerifyEmail(req.AccountId)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, RSuccess(""))
}

func apiAccountVerifyEmail(c *gin.Context) {
	req := &AccountVerifyEmailRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	account, err := accountcenter.Get().VerifyEmailCode(req.AccountId, req.Code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	//验证成功直接登录
	c.JSON(http.StatusOK, RSuccess(&AccountLoginResponse{
		AccountId: account.AccountId,
		Username:  account.Username,
		Type:      account.Type,
		Status:    account.Status,
		Token:     account.Token,
	}))
}

func apiCheckToken(c *gin.Context) {
	req := &CheckTokenRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	account, err := accountcenter.Get().CheckToken(req.AccountId, req.Token, req.AccType)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(&AccountLoginResponse{
		AccountId: account.AccountId,
		Username:  account.Username,
		Type:      account.Type,
		Status:    account.Status,
		Token:     account.Token,
	}))
}

func apiRequestRestPassword(c *gin.Context) {
	req := &RequestResetPasswordRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	err = accountcenter.Get().RequestResetPassword(req.Username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(""))
}

func apiResetPassword(c *gin.Context) {
	req := &ResetPasswordRequest{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	err = accountcenter.Get().ResetPassword(req.Username, req.Code, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(""))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "create", apiAccountCreate)
	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "login", apiAccountLogin)
	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "send_email", apiAccountSendVerifyEmail)
	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "verify_email", apiAccountVerifyEmail)
	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "check_token", apiCheckToken)

	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "request_reset_password", apiRequestRestPassword)
	web.GetService().RegisterApi(web.ApiModuleAccount, web.POST, "reset_password", apiResetPassword)
}
