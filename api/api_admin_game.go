package api

import (
	"dashfun_gamecenter/admin"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/gamecenter"
	"dashfun_gamecenter/web"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
)

type AdminGameImageType int

const (
	AdminGameImage_Icon AdminGameImageType = iota + 1
	AdminGameImage_Logo
	AdminGameImage_MainPic
)

type AdminGameDataRequest struct {
	Id     string                 `json:"id" form:"id"`
	Name   string                 `json:"name" form:"name"`
	Desc   string                 `json:"desc" form:"desc"`
	Url    string                 `json:"url" form:"url"`       //H5游戏部署地址
	Genre  []int                  `json:"genre" form:"genre"`   //游戏类型Id
	Status data.DashFunGameStatus `json:"status" form:"status"` //游戏状态
	//IconUrl    string `json:"iconUrl" form:"iconUrl"`       //游戏图标地址
	//LogoUrl    string `json:"logoUrl" form:"logoUrl"`       //游戏logo
	//MainPicUrl string `json:"mainPicUrl" form:"mainPicUrl"` //游戏主图地址 横向比例
}

type AdminGameSearchRequest struct {
	Keyword string                 `form:"keyword"`
	Genre   []int                  `json:"genre" form:"genre"` //游戏类型Id
	Status  data.DashFunGameStatus `json:"status" form:"status"`
	Size    int64                  `form:"size"`
	Page    int64                  `form:"page"`
}

type AdminGameUpdateImageRequest struct {
	Id string `json:"id" form:"id"` //游戏ID
}

// apiAdminGameCreate
//
//	@Summary	创建游戏
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		name	body		string									false	"游戏名称"
//	@Param		desc	body		string									false	"游戏介绍"
//	@Param		url		body		string									false	"游戏链接"
//	@Param		genre	body		[]int									false	"游戏类型"
//	@Success	200		{object}	api.JSONResult{data=[]data.DashFunGame}	"DashFunGame"
//	@Router		/api/v1/admin/game/create [post]
func apiAdminGameCreate(c *gin.Context, op *admin.AdminUser) {
	req := &AdminGameDataRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	if req.Name == "" || req.Desc == "" || req.Url == "" || len(req.Genre) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("missing required param"))
		return
	}

	game := gamecenter.Get().CreateGame(req.Name, req.Desc, req.Url, "", "", "", req.Genre)
	c.JSON(http.StatusOK, RSuccess(game))
}

// apiAdminUpdateUserStatus
//
//	@Summary	更新游戏
//	@Tags		Admin API
//	@Produce	json
//	@Accept		json
//	@Param		id		body		string									true	"要更新数据的游戏Id"
//	@Param		name	body		string									false	"游戏名称"
//	@Param		desc	body		string									false	"游戏介绍"
//	@Param		url		body		string									false	"游戏链接"
//	@Param		genre	body		[]int									false	"游戏类型"
//	@Success	200		{object}	api.JSONResult{data=[]data.DashFunGame}	"DashFunGame"
//	@Router		/api/v1/admin/game/update [post]
func apiAdminGameUpdate(c *gin.Context, op *admin.AdminUser) {
	req := &AdminGameDataRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	info, err := gamecenter.Get().UpdateGameInfo(req.Id, req.Name, req.Desc, req.Url, req.Genre, req.Status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, RSuccess(info))
}

// apiAdminGameUploadIcon
//
//	@Summary	上传游戏相关图片
//	@Tags		Admin API
//	@Produce	json
//	@Param		game_id	formData	string									true	"要更新数据的游戏Id"
//	@Param		icon	formData	file									false	"游戏图标，512x512"
//	@Param		logo	formData	file									false	"游戏logo"
//	@Param		main	formData	file									false	"游戏主图，640x320"
//	@Success	200		{object}	api.JSONResult{data=[]data.DashFunGame}	"DashFunGame"
//	@Router		/api/v1/admin/game/upload_image [post]
func apiAdminGameUploadImage(c *gin.Context, op *admin.AdminUser) {
	//req := &AdminGameUpdateImageRequest{}
	//if err := c.MultipartForm(req, nil); err != nil {
	//	c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
	//	return
	//}

	//if req.Id == "" {
	//	c.AbortWithStatusJSON(http.StatusBadRequest, RError("missing required param"))
	//	return
	//}

	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	paramId, ok := form.Value["game_id"]
	if !ok || len(paramId) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("form data game_id is required"))
		return
	}

	gameId := paramId[0]
	game, err := gamecenter.Get().FindGame(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	if game == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError("game not exist"))
		return
	}

	changed1, err := adminGameCheckAndUploadImage(form, game, AdminGameImage_Icon)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	changed2, err := adminGameCheckAndUploadImage(form, game, AdminGameImage_Logo)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	changed3, err := adminGameCheckAndUploadImage(form, game, AdminGameImage_MainPic)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}

	if changed1 || changed2 || changed3 {
		gamecenter.Get().SaveGame(game)
		c.JSON(http.StatusOK, RSuccess(game))
	}
}

// apiAdminGameSearch
//
//	@Summary	搜索游戏
//	@Tags		Admin API
//	@Accept		json
//	@Produce	json
//	@Param		keyword	body		string									false	"查询关键字"
//	@Param		genre	body		[]int									false	"查询类型"
//	@Param		status	body		data.DashFunGameStatus					false	"查询游戏状态"
//	@Param		size	body		int64									false	"每页数量"
//	@Param		page	body		int64									false	"当前页数，从1开始"
//	@Success	200		{object}	api.JSONResult{data=[]data.DashFunGame}	"Search Result"
//	@Router		/api/v1/admin/game/search [post]
func apiAdminGameSearch(c *gin.Context, op *admin.AdminUser) {
	req := &AdminGameSearchRequest{}
	if err := c.ShouldBindBodyWithJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, RError(err.Error()))
		return
	}
	games, totalPages, err := gamecenter.Get().FindGamesBackend(req.Keyword, req.Genre, req.Status, req.Size, req.Page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, RError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, PageSuccess(games, req.Page, req.Size, totalPages))
}

func getFormFileBytes(form *multipart.Form, name string) ([]byte, error) {
	data, ok := form.File[name]
	if ok {
		f := data[0]
		src, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()
		p := make([]byte, f.Size)
		_, err = src.Read(p)
		if err != nil {
			return nil, err
		}
		return p, nil
	}
	return nil, nil
}

// adminGameCheckAndUploadImage
// 从表单中读取数据，如果存在指定类型的图片数据，则上传到cos中，并更新到游戏信息中
func adminGameCheckAndUploadImage(form *multipart.Form, game *data.DashFunGame, imageType AdminGameImageType) (bool, error) {
	pName := ""
	switch imageType {
	case AdminGameImage_Icon:
		pName = "icon"
	case AdminGameImage_Logo:
		pName = "logo"
	case AdminGameImage_MainPic:
		pName = "main"
	}

	iconImg, ok := form.File[pName]
	if ok {
		imgFile := iconImg[0]
		src, err := imgFile.Open()
		if err != nil {
			return false, err
		}
		defer src.Close()
		p := make([]byte, imgFile.Size)
		_, err = src.Read(p)
		if err != nil {
			return false, err
		}
		switch imageType {
		case AdminGameImage_Icon:
			err := gamecenter.Get().UpdateGameIcon(game, p)
			if err != nil {
				return false, err
			}
		case AdminGameImage_Logo:
			err := gamecenter.Get().UpdateGameLogo(game, p)
			if err != nil {
				return false, err
			}
		case AdminGameImage_MainPic:
			err := gamecenter.Get().UpdateGameMainPic(game, p)
			if err != nil {
				return false, err
			}
		}

		return true, nil
	}
	return false, nil
}

func init() {
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "game/create", adminHandlerAuthWrapper(admin.AdminAuth_Game, apiAdminGameCreate))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "game/upload_image", adminHandlerAuthWrapper(admin.AdminAuth_Game, apiAdminGameUploadImage))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "game/update", adminHandlerAuthWrapper(admin.AdminAuth_Game, apiAdminGameUpdate))
	web.GetService().RegisterApi(web.ApiModuleAdmin, web.POST, "game/search", adminHandlerAuthWrapper(admin.AdminAuth_Game, apiAdminGameSearch))
}
