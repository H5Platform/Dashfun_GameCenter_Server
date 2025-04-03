package coincenter

import (
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/utils"
	"sync"
	"time"
)

// CoinUserDataList 用户持有coin数据缓存
type CoinUserDataList struct {
	sync.RWMutex
	// userId -> CoinsUserData
	coinsUserData map[string]*CoinsUserData
}

type CoinsUserData struct {
	sync.RWMutex
	userId  string
	coins   utils.List[*data.CoinUserData]
	records map[string]utils.List[*data.CoinUserRecordData] //用户coin变化记录，按照时间升序排列，从后往前取，key=coinId
}

func newCoinUserDataList() *CoinUserDataList {
	return &CoinUserDataList{
		coinsUserData: make(map[string]*CoinsUserData),
	}
}

func newCoinsUserData(userId string) *CoinsUserData {
	return &CoinsUserData{
		userId:  userId,
		coins:   utils.NewList[*data.CoinUserData](),
		records: make(map[string]utils.List[*data.CoinUserRecordData]),
	}
}

func newCoinUserRecordData(userId, coinId string, amount int32) *data.CoinUserRecordData {
	ret := &data.CoinUserRecordData{
		UserId: userId,
		CoinId: coinId,
		Change: amount,
		Time:   time.Now().UnixMilli(),
	}
	return ret
}

func newCoinUserData(userId, coinId string) *data.CoinUserData {
	return &data.CoinUserData{
		UserId:     userId,
		CoinId:     coinId,
		Amount:     0,
		CreateTime: time.Now().UnixMilli(),
	}
}

func (t *CoinUserDataList) Has(userId string) (*CoinsUserData, bool) {
	t.Lock()
	defer t.Unlock()
	d, exist := t.coinsUserData[userId]
	return d, exist
}

func (t *CoinUserDataList) GetCoinsUserData(userId string) *CoinsUserData {
	t.Lock()
	defer t.Unlock()
	d, exist := t.coinsUserData[userId]
	if !exist {
		//不存在则创建
		d = newCoinsUserData(userId)
		t.coinsUserData[userId] = d
	}
	return d
}

func (t *CoinUserDataList) RemoveCoinsUserData(userId string) {
	t.Lock()
	defer t.Unlock()
	_, exist := t.coinsUserData[userId]
	if exist {
		t.coinsUserData[userId] = nil
		delete(t.coinsUserData, userId)
	}
}

func (cud *CoinsUserData) AddOrUpdateUserData(coinUserData *data.CoinUserData) {
	for idx, item := range cud.coins.Items() {
		if item.CoinId == coinUserData.CoinId {
			//当前记录中包含这个任务的进度数据，做更新
			cud.coins.RemoveAt(idx)
			cud.coins.Add(coinUserData)
			return
		}
	}
	cud.coins.Add(coinUserData)
}

func (cud *CoinsUserData) GetCoinUserData(coinId string) *data.CoinUserData {
	for _, userData := range cud.coins.Items() {
		if userData.CoinId == coinId {
			return userData
		}
	}
	return nil
}

func (cud *CoinsUserData) AddUserCoinChangeRecord(record *data.CoinUserRecordData) {
	l := cud.GetUserCoinChangeRecordList(record.UserId)
	l.Add(record)
}

func (cud *CoinsUserData) GetUserCoinChangeRecordList(coinId string) utils.List[*data.CoinUserRecordData] {
	l, ok := cud.records[coinId]
	if !ok {
		l = utils.NewList[*data.CoinUserRecordData]()
		cud.records[coinId] = l
	}
	return l
}
