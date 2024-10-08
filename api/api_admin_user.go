package api

import (
	"dashfun_gamecenter/admin"
	"dashfun_gamecenter/admin/admin_user_mgr"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminCreateUserRequest struct {
	Username string              `json:"username" form:"username" binding:"required"`
	Email    string              `json:"email" form:"email" binding:"required"`
	Auth     admin.AdminUserAuth `json:"auth" form:"auth" binding:"required"`
}

type AdminUpdateUserRequest struct {
	UserId   string                `json:"user_id" form:"user_id" binding:"required"`
	Username string                `json:"username" form:"username"`
	Email    string                `json:"email" form:"email"`
	Auth     admin.AdminUserAuth   `json:"auth" form:"auth" `
	Status   admin.AdminUserStatus `json:"status" form:"status"`
}

type AdminUserResponse struct {
	UserId   string                `json:"user_id" form:"user_id" binding:"required"`
	Username string                `json:"username" form:"username"`
	Email    string                `json:"email" form:"email"`
	Auth     admin.AdminUserAuth   `json:"auth" form:"auth" `
	Status   admin.AdminUserStatus `json:"status" form:"status"`
}

func makeAdminUserResponse(user *admin.AdminUser) *AdminUserResponse {
	resp := &AdminUserResponse{
		UserId:   user.Id,
		Username: user.Name,
		Email:    user.Email,
		Auth:     user.Authorization,
		Status:   user.Status,
	}
	return resp
}

// apiAdminCreateUser
//
//	@Summary	创建后台账户
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		username	body		string									true	"用户名"
//	@Param		email		body		string									true	"邮箱"
//	@Param		auth		body		admin.AdminUserAuth						true	"权限"
//	@Success	200			{object}	api.JSONResult{data=AdminUserResponse}	"AdminUser"
//	@Router		/api/v1/admin/user/create [post]
func apiAdminCreateUser(c *gin.Context, op *admin.AdminUser) {
	req := &AdminCreateUserRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	u, err := admin_user_mgr.Get().CreateUser(req.Username, req.Email, req.Auth)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(makeAdminUserResponse(u)))
}

// apiAdminResetUserPassword
//
//	@Summary	重置用户密码
//	@Tags		Admin API
//	@Produce	json
//	@Param		user_id	formData	string									true	"用户ID"
//	@Success	200		{object}	api.JSONResult{data=AdminUserResponse}	"AdminUser"
//	@Router		/api/v1/admin/user/reset_password [post]
func apiAdminResetUserPassword(c *gin.Context, op *admin.AdminUser) {
	uid, ok := c.GetPostForm("user_id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("user_id is required"))
		return
	}

	targetUser, err := admin_user_mgr.Get().GetAdminUser(uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	err = admin_user_mgr.Get().ResetUserPassword(targetUser)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(makeAdminUserResponse(targetUser)))
}

// apiAdminUpdateUserBaseInfo
//
//	@Summary	更新用户信息
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		user_id		body		string									true	"用户ID"
//	@Param		email		body		string									true	"邮箱"
//	@Param		auth		body		admin.AdminUserAuth						true	"权限"
//	@Success	200			{object}	api.JSONResult{data=AdminUserResponse}	"admin user"
//	@Router		/api/v1/admin/user/update_base_info [post]
func apiAdminUpdateUserBaseInfo(c *gin.Context, op *admin.AdminUser) {
	req := &AdminUpdateUserRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	u, err := admin_user_mgr.Get().UpdateUserBaseInfo(req.UserId, req.Username, req.Email, req.Auth)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(makeAdminUserResponse(u)))
}

// apiAdminUpdateUserStatus
//
//	@Summary	修改用户状态
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		user_id	body		string									true	"用户ID"
//	@Param		status	body		int										true	"用户状态"
//	@Success	200		{object}	api.JSONResult{data=AdminUserResponse}	"admin user"
//	@Router		/api/v1/admin/user/update_status [post]
func apiAdminUpdateUserStatus(c *gin.Context, op *admin.AdminUser) {
	req := &AdminUpdateUserRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	u, err := admin_user_mgr.Get().UpdateUserStatus(req.UserId, req.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, RSuccess(makeAdminUserResponse(u)))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/create", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminCreateUser))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/reset_password", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminResetUserPassword))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/update_base_info", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminUpdateUserBaseInfo))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/update_status", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminUpdateUserStatus))
}
