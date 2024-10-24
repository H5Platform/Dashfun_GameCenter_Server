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

type AdminSearchUserRequest struct {
	Name   string                `json:"name" form:"name"`
	Email  string                `json:"email" form:"email"`
	Status admin.AdminUserStatus `json:"status" form:"status"`
	Page   int64                 `json:"page" form:"page"`
	Size   int64                 `json:"size" form:"size"`
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

// apiAdminActiveUser
//
//	@Summary	用户激活账户，设置密码
//	@Tags		Admin API
//	@Produce	json
//	@Param		new_password	formData	string									true	"用户新密码"
//	@Success	200				{object}	api.JSONResult{data=AdminUserResponse}	"AdminUser"
//	@Router		/api/v1/admin/user/active [post]
func apiAdminActiveUser(c *gin.Context, user *admin.AdminUser) {
	newPwd, ok := c.GetPostForm("new_password")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("new_password is required"))
		return
	}
	activeUser, err := admin_user_mgr.Get().ActiveUser(user, newPwd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(makeAdminUserResponse(activeUser)))
}

// apiAdminUpdateUserBaseInfo
//
//	@Summary	更新用户信息
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		user_id	body		string									true	"用户ID"
//	@Param		email	body		string									true	"邮箱"
//	@Param		auth	body		admin.AdminUserAuth						true	"权限"
//	@Success	200		{object}	api.JSONResult{data=AdminUserResponse}	"admin user"
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

// apiAdminUpdateUserStatus
//
//	@Summary	查询用户
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		name	body		string											false	"用户名称"
//	@Param		email	body		string											false	"用户email"
//	@Param		status	body		int												false	"用户状态"
//	@Param		size	body		int64											false	"每页数量"
//	@Param		page	body		int64											false	"当前页数，从1开始"
//	@Success	200		{object}	api.JSONResult{data=[]admin.AdminUserResult}	"Search Result"
//	@Router		/api/v1/admin/user/search [post]
func apiAdminSearchUsers(c *gin.Context, op *admin.AdminUser) {
	req := &AdminSearchUserRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	r, totalPages, err := admin_user_mgr.Get().GetUserList(req.Name, req.Email, req.Status, req.Size, req.Page)

	ar := make([]*admin.AdminUserResult, 0)

	for _, user := range r {
		ar = append(ar, admin.ToAdminUserResult(user))
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, PageSuccess(ar, req.Page, req.Size, totalPages))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/create", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminCreateUser))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/reset_password", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminResetUserPassword))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/update_base_info", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminUpdateUserBaseInfo))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/update_status", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminUpdateUserStatus))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "user/search", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminSearchUsers))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "/user/active", adminHandlerAuthWrapper(admin.AdminAuth_User, apiAdminActiveUser))
}
