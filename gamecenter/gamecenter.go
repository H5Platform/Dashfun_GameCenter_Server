package gamecenter

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/snowflake"
	"dashfun_gamecenter/tencentcos"
	"encoding/base64"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *GameCenter

type GameCenter struct {
	idGen     *snowflake.Worker
	secretGen *snowflake.Worker

	incOpenCountLock sync.Mutex
	gameLock         sync.RWMutex
	//游戏数据缓存
	id2games   map[string]*data.DashFunGame
	name2games map[string]*data.DashFunGame

	gameList     map[data.GameListType][]string
	gameListLock sync.RWMutex
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
	gc.gameLock = sync.RWMutex{}
	gc.incOpenCountLock = sync.Mutex{}

	gc.id2games = make(map[string]*data.DashFunGame)
	gc.name2games = make(map[string]*data.DashFunGame)

	//游戏列表缓存
	games, err := dao.GetGameDao().GetAllGames(0)
	if err != nil {
		log.Fatalf("GetGameDao.GetAllGames err:%v", err)
	}

	for _, game := range games {
		gc.processGame(game)
		gc.id2games[game.Id] = game
		gc.name2games[game.Name] = game
	}

	gc.gameList = make(map[data.GameListType][]string)
	gc.refreshGameList(data.GameListType_New)
	gc.refreshGameList(data.GameListType_Popular)
	gc.refreshGameList(data.GameListType_Suggest)
	gc.refreshGameList(data.GameListType_Banner)

	events.UserEnterGameEvents.On(gc.onUserEnterGame)

	//定时更新popular list
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			<-ticker.C
			gc.gameListLock.Lock()
			gc.sortPopularGameList()
			gc.gameListLock.Unlock()
		}
	}()
}

func (gc *GameCenter) newGameId() string {
	id := gc.idGen.NextId()
	return "gm" + strconv.FormatInt(id, 36)
}

func (gc *GameCenter) processGame(game *data.DashFunGame) {
	if game.ApiSecret == "" {
		//需要重新生成api secret
		game.ApiSecret = gc.genApiSecret()
		dao.GetGameDao().SaveOrUpdate(game)
	}
	_, existed := coincenter.Get().GetCoinByGame(game.Id)
	if !existed {
		_, err := coincenter.Get().CreateCoin("", game.Name, "", "", game.Id, false, 0, make(map[string]string))
		if err != nil {
			zap.S().Errorw("Auto CreateCoin failed", "gameId", game.Id, "err", err)
		}
	}
}

// CreateGame
// 创建一个游戏
func (gc *GameCenter) CreateGame(name, desc, url, iconUrl, logoUrl, mainPicUrl string, genre []int) (*data.DashFunGame, error) {
	f, err := gc.FindGameByName(name)
	if err != nil {
		return nil, err
	}
	if f != nil {
		return nil, errors.New("game name " + name + " already exists")
	}

	game := &data.DashFunGame{
		Id:          gc.newGameId(),
		Name:        name,
		Desc:        desc,
		Url:         url,
		IconUrl:     iconUrl,
		LogoUrl:     logoUrl,
		MainPicUrl:  mainPicUrl,
		Genre:       genre,
		Time:        time.Now().UnixMilli(),
		Status:      data.DashFunGameStatus_Pending,
		NewFlag:     0,
		PopularFlag: 0,
		SuggestFlag: 0,
		BannerFlag:  0,
		OpenCount:   0,
	}
	gc.processGame(game)
	gc.SaveGame(game)
	return game, nil
}

// updateGameImage
// 更新指定的游戏图片，上传到cos，并更新game数据
func (gc *GameCenter) updateGameImage(game *data.DashFunGame, img []byte, imgType string) error {
	imgName := ""

	uploadName := imgType + ".png"

	switch imgType {
	case "icon":
		imgName = game.IconUrl
	case "logo":
		imgName = game.LogoUrl
	case "mainPic":
		imgName = game.MainPicUrl
	}

	version := 1

	// 提取版本号
	if idx := strings.Index(imgName, "?v="); idx != -1 {
		if v, err := strconv.Atoi(imgName[idx+3:]); err == nil {
			version = v
		}
		imgName = imgName[:idx]
	}

	if imgName == "" || imgName != uploadName {
		imgName = imgType + ".png"
	}

	key := "images/" + game.Id + "/" + imgName
	_, err := tencentcos.Get().UploadData(key, img, "image/png")
	if err != nil {
		return err
	}

	imgName = imgName + "?v=" + strconv.Itoa(version+1)

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
// flags 四个数字， 1: new, 2: popular, 3: suggest, 4: banner ，每一个位置0表示关闭，>0表示排序值
func (gc *GameCenter) UpdateGameInfo(id, name, desc, url string, genre []int, openTime int64, status data.DashFunGameStatus, flags []int) (*data.DashFunGame, error) {
	game, err := gc.FindGame(id)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, errors.New("game not found")
	}

	update := false
	statusUpdated := false

	if status != data.DashFunGameStatus_NoChange {
		if status == data.DashFunGameStatus_Online {
			if game.IconUrl == "" || game.MainPicUrl == "" || game.LogoUrl == "" {
				return nil, errors.New("missing icon, logo or main picture")
			}
		}
		game.Status = status
		update = true
		statusUpdated = true
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

	changedList := make([]data.GameListType, 0)
	if len(flags) > 0 {
		oldFlgs := []int{game.NewFlag, game.PopularFlag, game.SuggestFlag, game.BannerFlag}
		game.NewFlag = 0
		game.PopularFlag = 0
		game.SuggestFlag = 0
		game.BannerFlag = 0
		for i, flag := range flags {
			switch i {
			case 0:
				game.NewFlag = flag
				if oldFlgs[0] != flag || statusUpdated {
					changedList = append(changedList, data.GameListType_New)
				}
			case 1:
				game.PopularFlag = flag
				if oldFlgs[1] != flag || statusUpdated {
					changedList = append(changedList, data.GameListType_Popular)
				}
			case 2:
				game.SuggestFlag = flag
				if oldFlgs[2] != flag || statusUpdated {
					changedList = append(changedList, data.GameListType_Suggest)
				}
			case 3:
				game.BannerFlag = flag
				if oldFlgs[3] != flag || statusUpdated {
					changedList = append(changedList, data.GameListType_Banner)
				}
			}
		}
		update = true
	}

	if update {
		_, err := gc.SaveGame(game)
		if err != nil {
			return nil, err
		}

		if len(changedList) > 0 {
			gc.gameListLock.Lock()
			defer gc.gameListLock.Unlock()
			for _, listType := range changedList {
				gc.refreshGameList(listType)
			}
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
	gc.gameLock.Lock()
	defer gc.gameLock.Unlock()
	gc.id2games[game.Id] = game
	gc.name2games[game.Name] = game

	return update, nil
}

func (gc *GameCenter) FindGamesById(gameIds ...string) []*data.DashFunGame {
	gc.gameLock.Lock()
	defer gc.gameLock.Unlock()

	var ret []*data.DashFunGame

	for _, gameId := range gameIds {
		var game *data.DashFunGame
		if gc.id2games[gameId] != nil {
			game = gc.id2games[gameId]
		} else {
			var err error
			game, err = dao.GetGameDao().GetGameById(gameId)
			if err != nil {
				continue
			}
			gc.id2games[game.Id] = game
			gc.name2games[game.Name] = game
		}
		if game != nil {
			ret = append(ret, game)
		}
	}
	return ret
}

func (gc *GameCenter) FindGame(gameId string) (*data.DashFunGame, error) {
	gc.gameLock.Lock()
	defer gc.gameLock.Unlock()
	var game *data.DashFunGame
	if gc.id2games[gameId] != nil {
		game = gc.id2games[gameId]
	} else {
		var err error
		game, err = dao.GetGameDao().GetGameById(gameId)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		gc.id2games[game.Id] = game
		gc.name2games[game.Name] = game

	}
	gc.processGame(game)
	return game, nil
}

func (gc *GameCenter) FindGameByName(gameName string) (*data.DashFunGame, error) {
	gc.gameLock.Lock()
	defer gc.gameLock.Unlock()
	var game *data.DashFunGame

	if gc.name2games[gameName] != nil {
		game = gc.name2games[gameName]
	} else {
		var err error
		game, err = dao.GetGameDao().GetGameByName(gameName)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
	}
	gc.processGame(game)
	return game, nil
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

	return dao.GetGameDao().FindGames(keyword, genre, nil, data.DashFunGameStatus_Online, size, page)
}

// sortPopularGameList 对popular的gamelist进行排序，其实不需要从数据库读取数据了，本身缓存数据已经可以对gamelist进行排序了
func (gc *GameCenter) sortPopularGameList() {
	gc.gameLock.RLock()
	defer gc.gameLock.RUnlock()

	games := make([]*data.DashFunGame, 0)
	for _, g := range gc.id2games {
		if g.Status == data.DashFunGameStatus_Online && (g.PopularFlag == 1 || g.OpenCount > 1) {
			games = append(games, g)
		}
	}

	slices.SortFunc(games, func(a, b *data.DashFunGame) int {
		if a.PopularFlag != b.PopularFlag {
			return b.PopularFlag - a.PopularFlag
		}
		if a.OpenCount != b.OpenCount {
			return b.OpenCount - a.OpenCount
		}
		return int(b.Time - a.Time)
	})

	list := make([]string, 0)
	for _, g := range games {
		list = append(list, g.Id)
	}
	gc.gameList[data.GameListType_Popular] = list
}

func (gc *GameCenter) refreshGameList(listType data.GameListType) {
	games, err := dao.GetGameDao().FindGameList(listType, 20)
	if err != nil {
		zap.S().Error("UpdateGameList "+strconv.Itoa(int(listType))+" failed", zap.Error(err))
		return
	}
	list := make([]string, 0)
	for _, game := range games {
		list = append(list, game.Id)
	}
	gc.gameList[listType] = list
	if listType == data.GameListType_Popular {
		//popular list需要和游戏开启次数统一排序
		gc.sortPopularGameList()
	}
}

// UpdateGameFlagBackend
// 后台更新游戏flag接口
//func (gc *GameCenter) UpdateGameFlagBackend(id string, newFlag, popularFlag, suggestFlag, bannerFlag int) {
//	game, err := gc.FindGame(id)
//	if err != nil {
//		return
//	}
//
//	changedList := make([]data.GameListType, 0)
//
//	if game.NewFlag != newFlag {
//		game.NewFlag = newFlag
//		changedList = append(changedList, data.GameListType_New)
//	}
//	if game.PopularFlag != popularFlag {
//		game.PopularFlag = popularFlag
//		changedList = append(changedList, data.GameListType_Popular)
//	}
//	if game.SuggestFlag != suggestFlag {
//		game.SuggestFlag = suggestFlag
//		changedList = append(changedList, data.GameListType_Suggest)
//	}
//	if game.BannerFlag != bannerFlag {
//		game.BannerFlag = bannerFlag
//		changedList = append(changedList, data.GameListType_Banner)
//	}
//
//	if len(changedList) > 0 {
//		gc.SaveGame(game)
//		for _, listType := range changedList {
//			gc.refreshGameList(listType)
//		}
//	}
//}

func (gc *GameCenter) GetGameList(listType data.GameListType) []string {
	gc.gameListLock.RLock()
	defer gc.gameListLock.RUnlock()
	return gc.gameList[listType]
}

// FindGamesBackend
// 游戏查询接口
// 后台使用
//
// params:
//
// keyword string 	-- 关键字，游戏名字包含
// genre []int		-- 游戏类型限定
// flags []int		-- 游戏标志限定
// status 			-- 游戏状态
// size int			-- 分页查询，每页数量
// page int			-- 分页查询，当前页数，从1开始
func (gc *GameCenter) FindGamesBackend(keyword string, genre, flags []int, status data.DashFunGameStatus, size int64, page int64) (games []*data.DashFunGame, total int, err error) {
	if page == 0 {
		page = 1
	}

	if genre == nil {
		genre = make([]int, 0)
	}

	if flags == nil {
		flags = make([]int, 0)
	}

	if size == 0 {
		size = 10
	}

	games, total, err = dao.GetGameDao().FindGames(keyword, genre, flags, status, size, page)

	if err == nil {
		for _, game := range games {
			gc.processGame(game)
		}
	}

	return
}

func (gc *GameCenter) GetGameGenres() []data.DashFunGameGenre {
	ret := make([]data.DashFunGameGenre, 0)
	for _, genre := range data.Genres {
		ret = append(ret, genre)
	}
	return ret
}

func (gc *GameCenter) genApiSecret() string {
	secret := "dashfun-" + gc.secretGen.NextStrId()
	secret = base64.StdEncoding.EncodeToString([]byte(secret))
	return secret
}

func (gc *GameCenter) onUserEnterGame(evt *events.EventUserEnterGame) {
	gc.incOpenCountLock.Lock()
	defer gc.incOpenCountLock.Unlock()
	game := evt.Game
	game.OpenCount++
	gc.SaveGame(game)
}
