package coincenter

import (
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/events"
	"dashfun_gamecenter/snowflake"
	"errors"
	"go.uber.org/zap"
	"log"
	"strconv"
	"strings"
	"sync"
)

var once sync.Once
var instance *CoinCenter

const (
	// CoinAddReasonRecalculate 手动重新计算用户的coin后，补偿丢失数据使用的Reason，通过record计算总额时，应去掉这个记录
	CoinAddReasonRecalculate = "Recalculate User Coin"
)

// CoinCenter 操作用户的coin
type CoinCenter struct {
	idGen       *snowflake.Worker
	coins       map[string]*data.CoinData
	coinsByName map[string]*data.CoinData
	users       *CoinUserDataList
}

func Get() *CoinCenter {
	once.Do(func() {
		instance = &CoinCenter{}
		instance.init()
	})
	return instance
}

func (c *CoinCenter) initDefaultCoins() {
	for _, cfg := range config.GetConfig().CoinCfg {
		_, exist := c.GetCoinByName(cfg.Name)
		if !exist {
			_, err := c.CreateCoin("", cfg.Name, cfg.Symbol, cfg.Desc, "", true, 100, make(map[string]string))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (c *CoinCenter) init() {
	c.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerCoinId))
	c.coins = make(map[string]*data.CoinData)
	c.coinsByName = make(map[string]*data.CoinData)

	c.users = newCoinUserDataList()

	coins, err := dao.GetCoinDao().GetAllCoins()
	if err != nil {
		log.Fatalf("GetCoinDao.GetAllCoins err:%v", err)
	}

	for _, coin := range coins {
		c.coins[coin.Id] = coin
		c.coinsByName[coin.Name] = coin
	}

	c.initDefaultCoins()

	events.UserLoginEvents.On(c.onUserLogin)
	events.UserPaymentEvents.On(c.onUserPayment)
}

// loadUserCoins 所有需要访问userCoins数据的方法，都需要用这个load先读取CoinsUserData，再进行操作，不要直接访问c.users
func (c *CoinCenter) loadUserCoins(userId string) (*CoinsUserData, error) {
	d, ok := c.users.Has(userId)
	if ok {
		return d, nil
	}

	coinsData, err := dao.GetCoinUserDao().GetAllUserCoins(userId)
	if err != nil {
		return nil, err
	}

	cud := c.users.CreateCoinsUserData(userId)
	cud.Lock()
	defer cud.Unlock()
	for _, coin := range coinsData {
		cud.AddOrUpdateUserData(coin)
		records, err := dao.GetCoinRecordDao().GetAllUserCoinRecords(userId, coin.CoinId)
		if err != nil {
			zap.S().Errorw("GetCoinRecordDao.GetAllUserCoinRecords err", "userId", userId, "coin", coin, "err", err)
			continue
		}

		for i := len(records) - 1; i >= 0; i-- {
			cud.AddUserCoinChangeRecord(records[i])
		}

	}

	return cud, nil
}

// recordUserCoinChange 记录用户coin数量变化，amount>0为增加 <0为扣减
func (c *CoinCenter) recordUserCoinChange(cud *CoinsUserData, coinId string, amount int32, reason, info string) {
	r := newCoinUserRecordData(cud.userId, coinId, amount, reason, info)
	_, err := dao.GetCoinRecordDao().AddRecord(r)
	if err != nil {
		zap.S().Errorw("save user coin change record error", "userId", cud.userId, "coinId", coinId, "amount", amount, "err", err)
	}
	cud.AddUserCoinChangeRecord(r)
}

func (c *CoinCenter) newCoinId() string {
	id := c.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (c *CoinCenter) GetAllCoins() []*data.CoinData {
	ret := make([]*data.CoinData, 0)
	for _, coin := range c.coins {
		ret = append(ret, coin)
	}
	return ret
}

func (c *CoinCenter) GetCoinById(coinId string) (*data.CoinData, bool) {
	coin, ok := c.coins[coinId]
	return coin, ok
}

func (c *CoinCenter) GetCoinByName(coinName string) (*data.CoinData, bool) {
	coin, ok := c.coinsByName[coinName]
	return coin, ok
}

func (c *CoinCenter) GetDashFunDiamond() *data.CoinData {
	coin, _ := c.GetCoinByName("DashFunDiamond")
	return coin
}

func (c *CoinCenter) GetDashFunXp() *data.CoinData {
	coin, _ := c.GetCoinByName("DashFunPoint")
	return coin
}

func (c *CoinCenter) GetDashFunTicket() *data.CoinData {
	coin, _ := c.GetCoinByName("DashFunTicket")
	return coin
}

func (c *CoinCenter) GetDashFunCoins() (ret []*data.CoinData) {
	for _, coin := range c.coins {
		if coin.BindGameId == "" || coin.BindGameId == "DashFun" {
			ret = append(ret, coin)
		}
	}
	return
}

func (c *CoinCenter) GetCoinByGame(gameId string) (*data.CoinData, bool) {
	for _, coin := range c.coins {
		if coin.BindGameId == gameId {
			return coin, true
		}
	}
	return nil, false
}

func (c *CoinCenter) CreateCoin(id, name, symbol, desc, bindGameId string, canWithdraw bool, minWithdraw float32, chainAddr map[string]string) (*data.CoinData, error) {
	if id == "" {
		id = c.newCoinId()
	}

	coin, err := dao.GetCoinDao().CreateCoin(id, name, symbol, desc, bindGameId, canWithdraw, minWithdraw, chainAddr)
	if err != nil {
		return nil, err
	}
	c.coins[coin.Id] = coin
	c.coinsByName[coin.Name] = coin
	return coin, nil
}

func (c *CoinCenter) UpdateCoin(id, name, desc, Symbol string, canWithdraw bool, minWithdraw float32, chainAddr map[string]string) (*data.CoinData, error) {
	coin, exist := c.GetCoinById(id)
	if !exist {
		return nil, errors.New("coin " + name + " not found")
	}
	if name != "" {
		coin.Name = name
	}
	if desc != "" {
		coin.Desc = desc
	}
	if Symbol != "" {
		coin.Symbol = Symbol
	}

	coin.CanWithdraw = canWithdraw
	if minWithdraw > 0 {
		coin.MinWithdraw = minWithdraw
	}
	if chainAddr != nil {
		coin.ChainAddr = chainAddr
	}
	dao.GetCoinDao().SaveOrUpdate(coin)
	return coin, nil
}

// AddUserCoinAmount 给用户增加指定数量的coin
func (c *CoinCenter) AddUserCoinAmount(userId, coinId string, amount int32, reason, info string) (*data.CoinUserData, error) {
	cud, err := c.loadUserCoins(userId)
	if err != nil {
		zap.S().Errorw("add user coin error, load user coins failed", "userId", userId, "coinId", coinId, "amount", amount, "reason", reason, "err", err)
		return nil, err
	}
	cud.Lock()
	defer cud.Unlock()

	coin, ok := c.GetCoinById(coinId)
	if !ok {
		zap.S().Errorw("add user coin error, coin not found", "userId", userId, "coinId", coinId, "amount", amount, "reason", reason)
		return nil, errors.New("coin not found")
	}
	coinData := cud.GetOrCreateCoinUserData(coinId) //c.GetCoinUserData(userId, coinId)
	if amount > 0 {
		coinData.Amount += amount
		cud.AddOrUpdateUserData(coinData)
		dao.GetCoinUserDao().SaveOrUpdate(coinData)
		c.recordUserCoinChange(cud, coinId, amount, reason, info)
		events.UserCoinChangedEvents.Emit(events.UserCoinChangedEvent{
			UserId: userId, Coin: coin, UserData: coinData, ChangedAmount: amount,
		})
		zap.S().Infow("add user coin amount", "userId", userId, "coin", coin, "amount", amount, "balance", coinData.Amount, "reason", reason, "info", info)
	}
	return coinData, nil
}

// DecUserCoinAmount 给用户减少指定数量的coin
func (c *CoinCenter) DecUserCoinAmount(userId, coinId string, amount int32, reason, info string) (*data.CoinUserData, error) {
	cud, err := c.loadUserCoins(userId)
	if err != nil {
		zap.S().Errorw("dec user coin error, load user coins failed", "userId", userId, "coinId", coinId, "amount", amount, "reason", reason, "err", err)
		return nil, err
	}

	cud.Lock()
	defer cud.Unlock()

	coin, ok := c.GetCoinById(coinId)
	if !ok {
		zap.S().Errorw("dec user coin error, coin not found", "userId", userId, "coinId", coinId, "amount", amount)
		return nil, errors.New("coin not found")
	}

	coinData := cud.GetOrCreateCoinUserData(coinId) //c.GetCoinUserData(userId, coinId)
	if amount > 0 {
		if coinData.Amount < amount {
			err := apperrors.ErrPaymentNotEnoughBalance
			zap.S().Errorw("dec user coin amount error", "userId", userId, "coin", coin, "amount", amount, "balance", coinData.Amount, "err", err)
			return nil, err
		}
		coinData.Amount -= amount
		cud.AddOrUpdateUserData(coinData)
		dao.GetCoinUserDao().SaveOrUpdate(coinData)
		events.UserCoinChangedEvents.Emit(events.UserCoinChangedEvent{
			UserId: userId, Coin: coin, UserData: coinData, ChangedAmount: -amount,
		})
		c.recordUserCoinChange(cud, coinId, -amount, reason, info)
		zap.S().Infow("dec user coin amount", "userId", userId, "coin", coin, "amount", amount, "balance", coinData.Amount, "reason", reason, "info", info)
	}
	return coinData, nil
}

// GetCoinUserData 获取用户指定coin的数据，外部使用，内部调用loadUserCoins，避免cud死锁
func (c *CoinCenter) GetCoinUserData(userId, coinId string) *data.CoinUserData {
	cud, err := c.loadUserCoins(userId)
	if err != nil {
		zap.S().Errorw("GetCoinUserData failed", "userId", userId, "coinId", coinId, "err", err)
		return nil
	}
	cud.Lock()
	defer cud.Unlock()
	cd := cud.GetOrCreateCoinUserData(coinId)
	return cd
}

// CalculateUserCoinByRecords 根据records计算用户coin的总额，返回增加、减少、余额
func (c *CoinCenter) CalculateUserCoinByRecords(userId, coinId string) (totalAdd, totalDec, totalBalance int32) {
	records := c.GetUserCoinRecords(userId, coinId, 0)
	for _, record := range records {
		if record.Reason == CoinAddReasonRecalculate {
			//CoinAddReasonRecalculate 不参与计算
			continue
		}
		if record.Change > 0 {
			totalAdd += record.Change
		} else if record.Change < 0 {
			totalDec += -record.Change
		}
		totalBalance += record.Change
	}
	return
}

// GetUserCoinRecords 获取用户指定coin的交易记录，count=0表示全部获取
func (c *CoinCenter) GetUserCoinRecords(userId, coinId string, count int) []*data.CoinUserRecordData {
	cud, err := c.loadUserCoins(userId)
	if err != nil {
		zap.S().Errorw("GetUserCoinRecords failed", "userId", userId, "coinId", coinId, "err", err)
		return nil
	}
	cud.RLock()
	defer cud.RUnlock()

	items := make([]*data.CoinUserRecordData, 0)

	if cud.records[coinId] != nil {
		items = cud.records[coinId].Items()
	}

	l := len(items)
	ret := make([]*data.CoinUserRecordData, 0)

	for i := l - 1; i >= 0; i-- {
		ret = append(ret, items[i])
		if count > 0 && len(ret) >= count {
			break
		}
	}
	return ret
}

func (c *CoinCenter) onUserPayment(evt *events.EventUserPayment) {
	if evt.Game.Id == "DashFun" && strings.HasPrefix(evt.Payment.Payload, "dashfun_buy_") {
		parts := strings.Split(evt.Payment.Payload, ":")
		if len(parts) < 2 {
			return
		}
		if parts[0] == "dashfun_buy_ticket" {
			amount, err := strconv.Atoi(parts[1])
			if err != nil {
				zap.S().Errorw("Invalid amount in payload", "payload", evt.Payment.Payload, "error", err)
				return
			}
			//购买DashFunTicket
			coin := c.GetDashFunTicket()
			_, err = c.AddUserCoinAmount(evt.User.Id, coin.Id, int32(amount), "Buy DashFunTicket", evt.Payment.Id)
			if err != nil {
				zap.S().Errorw("AddUser Ticket error", "userId", evt.User.Id, "coinId", coin.Id, "amount", amount, "error", err, "paymentId", evt.Payment.Id)
				return
			}
		}
	}
}
