package RechargeCenter

import (
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"errors"
	"github.com/stripe/stripe-go/v81"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *RechargeCenter

// RechargeCenter 充值中心
type RechargeCenter struct {
	idGen               *snowflake.Worker
	processingOrdersChn chan *data.DashFunRechargeData //已经支付完毕，还没有发放钻石的订单队列
}

func Get() *RechargeCenter {
	once.Do(func() {
		instance = &RechargeCenter{}
		instance.init()
	})
	return instance
}

func (r *RechargeCenter) init() {
	r.idGen = snowflake.Must(snowflake.GetWorker(data.WorkerRechargeId))
	//init stripe
	stripe.Key = config.GetConfig().StripeCfg.SecretKey
	r.processingOrdersChn = make(chan *data.DashFunRechargeData, 100)
	go r.processPaidOrders()
}

func (r *RechargeCenter) processPaidOrders() {
	diamond, _ := coincenter.Get().GetCoinByName("DashFunDiamond")
	for {
		order := <-r.processingOrdersChn
		//发放钻石
		_, err := coincenter.Get().AddUserCoinAmount(order.UserId, diamond.Id, int32(order.Diamond))
		if err != nil {
			zap.S().Errorw("AddUserCoinAmount failed", "order", order.Id, "err", err)
		}
		//更新order
		order.Status = data.DashFunRechargeStatus_Completed
		dao.GetRechargeDao().SaveOrUpdate(order)

		zap.S().Infow("recharge order completed", "order", order.Id, "price", order.Price, "diamond", order.Diamond)
	}
}

func (r *RechargeCenter) newRechargeOrderId() string {
	id := r.idGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (r *RechargeCenter) GetRechargePlatformOptions(platform string) RechargePlatformOptions {
	var options []RechargePlatformOption
	priceType := RechargePlatformOptionPriceTypeUSD

	//统一使用usd支付了
	//if platform == "ios" || platform == "android" {
	//	priceType = RechargePlatformOptionPriceTypeTGStar //目前ios和android都是tg的miniapp用户，使用星星支付
	//}

	for _, option := range config.GetConfig().RechargeCfg.Options {
		price := option.Price
		//暂时不用区分平台了，都用美元支付，而且是统一用浏览器支付，所以不需要区分是否是appstore的渠道
		//if platform == "ios" || platform == "android" { //目前ios和android都是tg的miniapp用户，使用星星支付
		//	price = option.TGStar
		//}
		options = append(options, RechargePlatformOption{
			Price:    price,
			Diamond:  option.Diamond,
			PriceOff: option.PriceOff,
		})
	}

	ret := RechargePlatformOptions{
		PriceType: priceType,
		Options:   options,
	}

	return ret
}

func (r *RechargeCenter) CreateRechargeOrder(userId string, rechargeOption config.RechargeOption, platform string, from data.RechargeFrom) (*data.DashFunRechargeData, error) {
	d := dao.GetRechargeDao()
	finalPrice := r.GetRechargePrice(rechargeOption, platform)
	recharge, err := d.CreateRecharge(r.newRechargeOrderId(), userId, from, finalPrice, rechargeOption.Diamond, "", "", time.Now().Unix())
	if err == nil {
		zap.S().Infow("CreateRechargeOrder", "recharge", recharge)
	}
	return recharge, err
}

func (r *RechargeCenter) GetRechargePrice(rechargeOption config.RechargeOption, platform string) int {
	price := rechargeOption.Price
	//暂时不用区分平台了，都用美元支付，而且是统一用浏览器支付，所以不需要区分是否是appstore的渠道
	finalPrice := price
	if rechargeOption.PriceOff > 0 {
		finalPrice = price * (1000 - rechargeOption.PriceOff) / 1000
	}
	return finalPrice
}

func (r *RechargeCenter) GetRechargeOrder(id string) (*data.DashFunRechargeData, error) {
	order, err := dao.GetRechargeDao().FindRechargeById(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrRechargeOrderNotFound
		} else {
			return nil, err
		}
	}
	return order, nil
}

func (r *RechargeCenter) PendingRechargeOrder(orderId string, payFrom string) (*data.DashFunRechargeData, error) {
	order, err := r.GetRechargeOrder(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status != data.DashFunRechargeStatus_Created {
		return nil, apperrors.ErrRechargeOrderStatus
	}

	order.Status = data.DashFunRechargeStatus_Pending
	order.PayFrom = payFrom

	return dao.GetRechargeDao().SaveOrUpdate(order)
}

func (r *RechargeCenter) ConfirmRechargeOrder(orderId string, channelOrderId string) (*data.DashFunRechargeData, error) {
	order, err := r.GetRechargeOrder(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status != data.DashFunRechargeStatus_Pending {
		return nil, apperrors.ErrRechargeOrderStatus
	}

	order.ChannelPayId = channelOrderId
	order.Status = data.DashFunRechargeStatus_Paid
	order.PaidAt = time.Now().Unix()

	order, err = dao.GetRechargeDao().SaveOrUpdate(order)
	if err != nil {
		return nil, err
	}

	//发送订单到处理队列
	r.processingOrdersChn <- order
	return order, nil
}

func (r *RechargeCenter) CancelRechargeOrder(orderId, userId string) (*data.DashFunRechargeData, error) {
	order, err := r.GetRechargeOrder(orderId)
	if err != nil {
		return nil, err
	}

	if order.Status > data.DashFunRechargeStatus_Pending {
		return nil, apperrors.ErrRechargeOrderStatus
	}

	if order.UserId != userId {
		return nil, apperrors.ErrRechargeOrderCantCancel
	}

	order.Status = data.DashFunRechargeStatus_Canceled
	order.Message = "Canceled Manually"
	order.PaidAt = time.Now().Unix()

	return dao.GetRechargeDao().SaveOrUpdate(order)
}
