package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/fishingverse"
	"dashfun_gamecenter/usercenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

func apiUserUpdateProfile(c *gin.Context, user *data.DashFunUser) {

	nickname, existed := c.GetPostForm("nickname")
	if !existed {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("nickname is required"))
		return
	}

	avatar, err := c.FormFile("avatar.png")
	avt := make([]byte, 0)
	if err == nil && avatar != nil {
		f, err := avatar.Open()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, RError("failed to open avatar file"))
			return
		}
		defer f.Close()
		avt, err = io.ReadAll(f)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, RError("failed to read avatar file"))
			return
		}
	}

	u, err := usercenter.Get().UserUpdateProfile(user.Id, nickname, avt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(&data.UserProfileData{
		UserId:   u.Id,
		Nickname: u.Nickname,
		Avatar:   u.AvatarUrl,
	}))
}

func apiGetUserDailyCheckInRemaining(c *gin.Context, user *data.DashFunUser) {
	remaining := fishingverse.Get().GetUserDailyPostRemaining(user.Id)
	c.JSON(http.StatusOK, RSuccess(remaining))
}

func apiUserPost(c *gin.Context, user *data.DashFunUser) {
	postContent, existed := c.GetPostForm("post")
	if !existed || postContent == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("post content is required"))
		return
	}

	location, _ := c.GetPostForm("location")
	fish, _ := c.GetPostForm("fish")

	postId, err := fishingverse.Get().Post(user.Id, postContent, location, fish)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(postId))
}

func apiGetPosts(c *gin.Context, user *data.DashFunUser) {
	limitStr := c.Query("limit")
	limit := 50 // default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, RError("invalid limit"))
			return
		}
	}

	posts, err := fishingverse.Get().GetPosts(limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(posts))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleFishingVerse, web.POST, "profile/update", userHandlerAuthWrapper(apiUserUpdateProfile))
	web.GetService().RegisterApi(web.ApiModuleFishingVerse, web.GET, "remaining", userHandlerAuthWrapper(apiGetUserDailyCheckInRemaining))
	web.GetService().RegisterApi(web.ApiModuleFishingVerse, web.POST, "post", userHandlerAuthWrapper(apiUserPost))
	web.GetService().RegisterApi(web.ApiModuleFishingVerse, web.GET, "posts", userHandlerAuthWrapper(apiGetPosts))
}
