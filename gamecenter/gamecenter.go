package gamecenter

import (
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tencentcos"
	"encoding/base64"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *GameCenter

type GameCenter struct {
	idGen     *snowflake.Worker
	secretGen *snowflake.Worker
}

func Get() *GameCenter {
	once.Do(func() {
		instance = &GameCenter{}
		instance.init()
	})
	return instance
}

func (gc *GameCenter) init() {
	gc.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerGameId))
	gc.secretGen = snowflake.Must(snowflake.GetWorker(data.WorkerApiSecret))
}

func (gc *GameCenter) newGameId() string {
	id := gc.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (gc *GameCenter) processGame(game *data.DashFunGame) {
	if game.ApiSecret == "" {
		//需要重新生成api secret
		game.ApiSecret = gc.genApiSecret()
		dao.GetGameDao().SaveOrUpdate(game)
	}
}

// CreateGame
// 创建一个游戏
func (gc *GameCenter) CreateGame(name, desc, url, iconUrl, logoUrl, mainPicUrl string, genre []int) *data.DashFunGame {
	game := &data.DashFunGame{
		Id:         gc.newGameId(),
		Name:       name,
		Desc:       desc,
		Url:        url,
		IconUrl:    iconUrl,
		LogoUrl:    logoUrl,
		MainPicUrl: mainPicUrl,
		Genre:      genre,
		Time:       time.Now().UnixMilli(),
		Status:     data.DashFunGameStatus_Pending,
	}
	gc.processGame(game)
	return game
}

// updateGameImage
// 更新指定的游戏图片，上传到cos，并更新game数据
func (gc *GameCenter) updateGameImage(game *data.DashFunGame, img []byte, imgType string) error {
	imgName := ""
	switch imgType {
	case "icon":
		imgName = game.IconUrl
	case "logo":
		imgName = game.LogoUrl
	case "mainPic":
		imgName = game.MainPicUrl
	}

	if imgName == "" {
		imgName = tencentcos.Get().NextName() + ".png"
	}
	key := "images/" + game.Id + "/" + imgName
	_, err := tencentcos.Get().UploadData(key, img, "image/png")
	if err != nil {
		return err
	}

	switch imgType {
	case "icon":
		game.IconUrl = imgName
	case "logo":
		game.LogoUrl = imgName
	case "mainPic":
		game.MainPicUrl = imgName
	}

	_, err = gc.SaveGame(game)
	if err != nil {
		return err
	}
	return nil
}

// UpdateGameInfo 更新游戏信息，如果openTime不需要变更，传入-1
func (gc *GameCenter) UpdateGameInfo(id, name, desc, url string, genre []int, openTime int64, status data.DashFunGameStatus) (*data.DashFunGame, error) {
	game, err := gc.FindGame(id)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, errors.New("game not found")
	}

	update := false

	if status != data.DashFunGameStatus_NoChange {
		if status == data.DashFunGameStatus_Online {
			if game.IconUrl == "" || game.MainPicUrl == "" || game.LogoUrl == "" {
				return nil, errors.New("missing icon, logo or main picture")
			}
		}
		game.Status = status
		update = true
	}

	if name != "" {
		game.Name = name
		update = true
	}
	if desc != "" {
		game.Desc = desc
		update = true
	}
	if url != "" {
		game.Url = url
		update = true
	}
	if openTime >= 0 {
		game.OpenTime = openTime
		update = true
	}
	if genre != nil && len(genre) > 0 {
		game.Genre = genre
		update = true
	}

	if update {
		_, err := gc.SaveGame(game)
		if err != nil {
			return nil, err
		}
	}
	return game, nil
}

// UpdateGameLogo
// 更新游戏的Logo， 上传到指定的cos，更新游戏的logoUrl属性
func (gc *GameCenter) UpdateGameLogo(game *data.DashFunGame, logo []byte) error {
	return gc.updateGameImage(game, logo, "logo")
}

// UpdateGameMainPic
// 更新游戏的MainPic， 上传到指定的cos，更新游戏的mainPicUrl属性
func (gc *GameCenter) UpdateGameMainPic(game *data.DashFunGame, mainPic []byte) error {
	return gc.updateGameImage(game, mainPic, "mainPic")
}

// UpdateGameIcon
// 更新游戏的图标， 上传到指定的cos，更新游戏的iconUrl属性
func (gc *GameCenter) UpdateGameIcon(game *data.DashFunGame, icon []byte) error {
	return gc.updateGameImage(game, icon, "icon")
}

func (gc *GameCenter) ChangeGameStatus(gameId string, status data.DashFunGameStatus) (*data.DashFunGame, error) {
	game, err := gc.FindGame(gameId)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, errors.New("game not found")
	}

	game.Status = status
	gc.SaveGame(game)

	return game, nil
}

func (gc *GameCenter) SaveGame(game *data.DashFunGame) (*data.DashFunGame, error) {
	if game.Id == "" {
		game.Id = gc.newGameId()
	}
	update, err := dao.GetGameDao().SaveOrUpdate(game)
	if err != nil {
		return nil, err
	}
	return update, nil
}

func (gc *GameCenter) FindGame(gameId string) (*data.DashFunGame, error) {
	game, err := dao.GetGameDao().GetGameById(gameId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	gc.processGame(game)
	return game, err
}

func (gc *GameCenter) FindGameByName(gameName string) (*data.DashFunGame, error) {
	game, err := dao.GetGameDao().GetGameByName(gameName)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	gc.processGame(game)
	return game, err
}

// FindGames
// 游戏查询接口
// 前台使用，只查询已经上线的游戏
//
// params:
//
// keyword string 	-- 关键字，游戏名字包含
// genre []int		-- 游戏类型限定
// size int64			-- 分页查询，每页数量
// page int64			-- 分页查询，当前页数，从1开始
func (gc *GameCenter) FindGames(keyword string, genre []int, size int64, page int64) (games []*data.DashFunGame, totalPages int, err error) {
	if page == 0 {
		page = 1
	}

	if genre == nil {
		genre = make([]int, 0)
	}

	if size == 0 {
		size = 10
	}

	return dao.GetGameDao().FindGames(keyword, genre, data.DashFunGameStatus_Online, size, page)

}

// FindGamesBackend
// 游戏查询接口
// 后台使用
//
// params:
//
// keyword string 	-- 关键字，游戏名字包含
// genre []int		-- 游戏类型限定
// status 			-- 游戏状态
// size int			-- 分页查询，每页数量
// page int			-- 分页查询，当前页数，从1开始
func (gc *GameCenter) FindGamesBackend(keyword string, genre []int, status data.DashFunGameStatus, size int64, page int64) (games []*data.DashFunGame, total int, err error) {
	if page == 0 {
		page = 1
	}

	if genre == nil {
		genre = make([]int, 0)
	}

	if size == 0 {
		size = 10
	}

	return dao.GetGameDao().FindGames(keyword, genre, status, size, page)
}

func (gc *GameCenter) genApiSecret() string {
	secret := "dashfun-" + gc.secretGen.NextStrId()
	secret = base64.StdEncoding.EncodeToString([]byte(secret))
	return secret
}

func init() {

}
