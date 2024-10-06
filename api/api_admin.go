package api

import (
	"dashfun_gamecenter/admin"
	"dashfun_gamecenter/admin/admin_user_mgr"
	"dashfun_gamecenter/web"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"net/http"
	"strings"
)

type AdminLoginData struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type AdminAuthData struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

// apiAdminUserLogin
//	@Router	/api/v1/admin/login [post]
func apiAdminUserLogin(c *gin.Context) {
	d := &AdminLoginData{}
	err := c.ShouldBindBodyWithJSON(d)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	_, ali, err := admin_user_mgr.Get().Login(d.Username, d.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(ali))
}

// checkAdminAuth
// 验证admin的信息，Bearer + Base64({user_id:userId, token: loginToken})
func checkAdminUserLogin(c *gin.Context) (*admin.AdminUser, error) {
	authData := c.GetHeader("authorization")
	if len(authData) == 0 || !strings.HasPrefix(authData, "Bearer ") {
		return nil, errors.New("unauthorized")
	}
	s := strings.Split(authData, "Bearer ")[1]

	decodeString, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	data := &AdminAuthData{}

	err = json.Unmarshal(decodeString, data)
	if err != nil {
		return nil, err
	}

	au, auth := admin_user_mgr.Get().CheckToken(data.UserId, data.Token)

	if !auth {
		return nil, errors.New("unauthorized")
	}

	return au, nil
}

func checkAdminUserAuth(user *admin.AdminUser, authRequired admin.AdminUserAuth) bool {
	if user.Authorization == admin.AdminAuth_Admin {
		return true
	}

	if user.Authorization&authRequired == authRequired {
		return true
	}

	return false
}

func checkAdminAuth(c *gin.Context, authRequired admin.AdminUserAuth) (*admin.AdminUser, error) {
	au, err := checkAdminUserLogin(c)
	if err != nil {
		return nil, err
	}

	hasAuth := checkAdminUserAuth(au, admin.AdminAuth_User)
	if !hasAuth {
		return nil, errors.New("unauthorized")
	}

	return au, nil
}

func adminHandlerAuthWrapper(authRequired admin.AdminUserAuth, handler func(ctx *gin.Context, user *admin.AdminUser)) func(*gin.Context) {
	return func(c *gin.Context) {
		au, err := checkAdminAuth(c, authRequired)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, RError(err.Error()))
			return
		}
		if handler != nil {
			handler(c, au)
		}
	}
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "login", apiAdminUserLogin)
}
