package api

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/nolandev"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func apiNolanGetUserDailyCheckinRemaining(c *gin.Context, user *data.DashFunUser) {
	remaining := nolandev.Get().GetUserDailyPostRemaining(user.Id)
	c.JSON(http.StatusOK, RSuccess(remaining))
}

func apiNolanUserPost(c *gin.Context, user *data.DashFunUser) {
	postContent, existed := c.GetPostForm("post")
	if !existed || postContent == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("post content is required"))
		return
	}

	location, _ := c.GetPostForm("location")
	fish, _ := c.GetPostForm("fish")

	postId, err := nolandev.Get().Post(user.Id, postContent, location, fish)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(postId))
}

func apiNolanGetPosts(c *gin.Context, user *data.DashFunUser) {
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

	posts, err := nolandev.Get().GetPosts(limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(posts))
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleNolanDev, web.POST, "profile/update", userHandlerAuthWrapper(apiUserUpdateProfile))
	web.GetService().RegisterApi(web.ApiModuleNolanDev, web.GET, "remaining", userHandlerAuthWrapper(apiNolanGetUserDailyCheckinRemaining))
	web.GetService().RegisterApi(web.ApiModuleNolanDev, web.POST, "post", userHandlerAuthWrapper(apiNolanUserPost))
	web.GetService().RegisterApi(web.ApiModuleNolanDev, web.GET, "posts", userHandlerAuthWrapper(apiNolanGetPosts))
}
