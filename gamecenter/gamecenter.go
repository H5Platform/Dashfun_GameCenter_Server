package gamecenter

import (
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"encoding/base64"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"sync"
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

func (gc *GameCenter) genApiSecret() string {
	secret := "dashfun-" + gc.secretGen.NextStrId()
	secret = base64.StdEncoding.EncodeToString([]byte(secret))
	return secret
}

func init() {

}
