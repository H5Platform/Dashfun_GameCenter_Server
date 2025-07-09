package AirdropCenter

import (
	"context"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/utils"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"math"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *AirdropCenter

type VestingRequestStatus int

const (
	VestingStatusCreated VestingRequestStatus = iota + 1 // 锁仓请求已创建
	VestingStatusPending
	VestingStatusDone
	VestingStatusFailed
)

type CreateVestingRequest struct {
	UserId           string               `json:"user_id"`            // 用户ID
	Address          string               `json:"address"`            // 用户提交的领取地址
	TokenAmount      string               `json:"token_amount"`       // 代币数量，单位为ether
	IgnoreInitUnlock bool                 `json:"ignore_init_unlock"` // 是否忽略初始解锁（KuCoin用户不需要初始解锁）
	KcUid            string               `json:"kc_uid"`             // KuCoin ID，用户自行绑定
	Status           VestingRequestStatus `json:"status"`             // 锁仓请求状态
	Time             time.Time            `json:"time"`               // 状态变化时间
	Result           string               `json:"result"`             // 交易结果，tx hash 或错误信息
	AirdropData      *data.AirdropData    `json:"-"`                  // Airdrop数据
}

type KuCoinUserDetail struct {
	KcUid       string `json:"kc_uid"`
	TotalAmount string `json:"total_amount"`
	Address     string `json:"address"`
}

type AirdropCenter struct {
	requestList utils.List[*CreateVestingRequest]        // 锁仓请求列表
	requestMap  utils.Map[string, *CreateVestingRequest] // 锁仓请求映射，key为用户ID
	scheduler   *utils.Scheduler
}

func Get() *AirdropCenter {
	once.Do(func() {
		instance = &AirdropCenter{}
		instance.init()
	})

	return instance
}

func (a *AirdropCenter) init() {
	a.requestList = utils.NewSynchronizedList[*CreateVestingRequest]()
	a.requestMap = utils.NewSynchronizedMap[string, *CreateVestingRequest]()
	a.scheduler = utils.NewScheduler(context.TODO())
	a.scheduler.Add("airdrop_requests_schedule", time.Second, a.scheduleRequests)
}

func (a *AirdropCenter) GetAllKuCoinUsersDetail() ([]*KuCoinUserDetail, error) {
	d := dao.GetAirdropDao()
	airdropData, err := d.GetAllAirdropData()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all airdrop data")
	}
	var details []*KuCoinUserDetail

	kcUidSet := make(map[string]*KuCoinUserDetail)

	for _, ad := range airdropData {
		if ad.KuCoinId != "" && ad.ClaimedTime > 0 {
			key := ad.KuCoinId
			if _, ok := kcUidSet[key]; ok {
				// 如果已经存在这个KuCoin ID，则累加总金额
				kcUidSet[key].TotalAmount = fmt.Sprintf("%.8f", toFloat64(kcUidSet[key].TotalAmount)+toFloat64(ad.TokenAmount))
			} else {
				// 如果不存在这个KuCoin ID，则创建一个新的记录
				details = append(details, &KuCoinUserDetail{
					KcUid:       ad.KuCoinId,
					TotalAmount: ad.TokenAmount,
					Address:     ad.WalletAddress,
				})
				kcUidSet[key] = details[len(details)-1]
			}
		}
	}
	return details, nil
}

func (a *AirdropCenter) getUserTokenAmount(userId string) string {
	cfg := config.GetConfig().AirdropCfg

	xpCoin := coincenter.Get().GetDashFunXp()
	coinData := coincenter.Get().GetCoinUserData(userId, xpCoin.Id)
	proportion := float64(coinData.Amount) / float64(cfg.TotalScore)
	tokenAmount := math.Floor(float64(cfg.TotalToken)*proportion*1e8) / 1e8 //分到的token数量，保留8位小数 //分到的token数量

	amtString := "0"
	if tokenAmount > 0 {
		amtString = fmt.Sprintf("%.8f", tokenAmount)
	}
	return amtString
}

func (a *AirdropCenter) GetUserCreateVestingRequest(userId string) (*CreateVestingRequest, bool) {
	req, ok := a.requestMap.Get(userId)
	return req, ok
}

func (a *AirdropCenter) GetAirdropUserDetail(userId string) (*AirdropUserDetail, error) {
	d := dao.GetAirdropDao()
	cfg := config.GetConfig().AirdropCfg
	airdropData, err := d.GetAirdropData(userId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			amtString := a.getUserTokenAmount(userId)
			airdropData = &data.AirdropData{
				UserId:        userId,
				KuCoinId:      "",
				TokenAmount:   amtString,
				ClaimedTime:   0,
				WalletAddress: "",
			}
			d.SaveOrUpdate(airdropData)
		} else {
			return nil, err
		}
	}

	if airdropData.ClaimedTime == 0 && time.Now().Unix() < cfg.GetLockXpTime() {
		// 如果还没有领取，并且锁定积分时间还没到，则更新积分
		amtString := a.getUserTokenAmount(userId)
		if airdropData.TokenAmount != amtString {
			airdropData.TokenAmount = amtString
			d.SaveOrUpdate(airdropData)
		}
	}

	claimed := false
	if airdropData.ClaimedTime > 0 {
		claimed = true
	}

	detail := &AirdropUserDetail{
		StartTime:       cfg.GetStartTime(),
		ClaimTime:       int64(cfg.ClaimTime),
		TokenAmount:     airdropData.TokenAmount,
		Claimed:         claimed,
		VestingContract: cfg.VestingContract,
		TokenContract:   cfg.TokenContract,
		ClaimAddress:    airdropData.WalletAddress,
		KuCoinId:        airdropData.KuCoinId,
	}

	return detail, nil

}

func (a *AirdropCenter) CreateVestingForUser_(userId string, address string, kcUid string) (string, error) {
	cfg := config.GetConfig().AirdropCfg
	if cfg.VestingContract == "" {
		return "", errors.New("vesting contract is not configured")
	}

	detail, err := a.GetAirdropUserDetail(userId)
	if err != nil {
		return "", errors.Wrap(err, "failed to get airdrop user detail")
	}

	if detail.Claimed {
		return "", errors.New("you have already claimed your airdrop")
	}

	d := dao.GetAirdropDao()
	airdropInfo, err := d.GetAirdropData(userId)
	if err != nil {
		return "", errors.Wrap(err, "failed to get airdrop data")
	}

	vc, err := NewVestingContract(cfg.VestingContract)
	if err != nil {
		return "", errors.Wrap(err, "failed to create vesting contract")
	}

	ignoreInitUnlock := false
	if kcUid != "" {
		ignoreInitUnlock = true // KuCoin users do not need to unlock the initial amount
	}
	tx, err := vc.CreateVesting(address, detail.TokenAmount, ignoreInitUnlock)
	if err != nil {
		return "", errors.Wrap(err, "failed to create vesting transaction")
	}

	airdropInfo.WalletAddress = address
	airdropInfo.KuCoinId = kcUid
	airdropInfo.ClaimedTime = time.Now().Unix()
	airdropInfo.ClaimTxHash = tx

	d.SaveOrUpdate(airdropInfo)

	zap.S().Infow("Airdrop claimed", "userId", userId, "address", address, "kc_uid", kcUid, "tx", tx)

	return tx, nil
}

func (a *AirdropCenter) CreateVestingForUser(userId string, address string, kcUid string) (string, error) {
	cfg := config.GetConfig().AirdropCfg
	if cfg.VestingContract == "" {
		return "", errors.New("vesting contract is not configured")
	}
	d := dao.GetAirdropDao()
	airdropInfo, err := d.GetAirdropData(userId)
	if err != nil {
		return "", errors.Wrap(err, "failed to get airdrop data")
	}

	detail, err := a.GetAirdropUserDetail(userId)
	if err != nil {
		return "", errors.Wrap(err, "failed to get airdrop user detail")
	}

	if detail.Claimed {
		return "", errors.New("you have already claimed your airdrop")
	}

	if _, ok := a.GetUserCreateVestingRequest(userId); ok {
		return "", errors.New("your claim request is already in progress")
	}

	ignoreInitUnlock := false
	if kcUid != "" {
		ignoreInitUnlock = true // KuCoin users do not need to unlock the initial amount
	}

	req := &CreateVestingRequest{
		UserId:           userId,
		Address:          address,
		TokenAmount:      detail.TokenAmount,
		IgnoreInitUnlock: ignoreInitUnlock,
		KcUid:            kcUid,
		Status:           VestingStatusCreated,
		Result:           "",
		AirdropData:      airdropInfo,
		Time:             time.Now(),
	}

	airdropInfo.WalletAddress = address
	airdropInfo.KuCoinId = kcUid
	airdropInfo.ClaimedTime = time.Now().Unix()
	d.SaveOrUpdate(airdropInfo)

	a.requestList.Add(req)
	a.requestMap.Set(userId, req)

	return "created", nil // 返回一个标识，表示请求已创建，实际的处理会在调度器中进行
}

func (a *AirdropCenter) makeRequestFailed(request *CreateVestingRequest, reason string) {
	request.Status = VestingStatusFailed
	request.Result = reason
	request.Time = time.Now()
	zap.S().Errorw("Airdrop request failed", "userId", request.UserId, "address", request.Address, "reason", reason)
	//从队列中移除request
	a.requestList.Remove(request)
}

func (a *AirdropCenter) makeRequestDone(request *CreateVestingRequest, txHash string) {
	request.Status = VestingStatusDone
	request.Result = txHash
	request.Time = time.Now()
	zap.S().Infow("Airdrop request done", "userId", request.UserId, "address", request.Address, "txHash", txHash)
	a.requestList.Remove(request)
}

func (a *AirdropCenter) DoTransaction(request *CreateVestingRequest) {
	cfg := config.GetConfig().AirdropCfg
	address := request.Address
	tokenAmount := request.TokenAmount
	ignoreInitUnlock := request.IgnoreInitUnlock
	kcUid := request.KcUid
	airdropInfo := request.AirdropData

	vc, err := NewVestingContract(cfg.VestingContract)
	if err != nil {
		a.makeRequestFailed(request, "Failed:"+err.Error())
		return
	}
	tx, err := vc.CreateVesting(address, tokenAmount, ignoreInitUnlock)
	if err != nil {
		a.makeRequestFailed(request, "Failed:"+err.Error())
		return
	}

	airdropInfo.WalletAddress = address
	airdropInfo.KuCoinId = kcUid
	airdropInfo.ClaimedTime = time.Now().Unix()
	airdropInfo.ClaimTxHash = tx

	a.makeRequestDone(request, tx)

	d := dao.GetAirdropDao()
	d.SaveOrUpdate(airdropInfo)
}

func (a *AirdropCenter) scheduleRequests() {
	if a.requestList.Size() == 0 {
		return // No requests to process
	}

	req := a.requestList.Get(0)
	if req.Status == VestingStatusPending {
		distance := time.Since(req.Time)
		if distance > time.Minute*2 {
			// 如果请求已经超过2分钟没有处理，则标记为失败
			a.makeRequestFailed(req, "Request timed out after 5 minutes")
		}
		return
	}

	if req.Status != VestingStatusCreated {
		//不是pending也不是created状态的请求，移除
		a.requestList.Remove(req)
		return
	}

	// 将请求状态设置为Pending，表示正在处理
	req.Status = VestingStatusPending
	req.Time = time.Now()
	go a.DoTransaction(req)
}

func toFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
