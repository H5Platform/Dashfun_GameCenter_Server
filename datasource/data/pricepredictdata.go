package data

import (
	"fmt"
	"time"
)

type PricePredictStatus int

const (
	PricePredictStatusUnsubmitted PricePredictStatus = 0 // 未提交
	PricePredictStatusPending     PricePredictStatus = 1 // 待开奖
	PricePredictStatusRevealed    PricePredictStatus = 2 // 已开奖
	PricePredictStatusClaimed     PricePredictStatus = 3 // 已领取
)

type PricePredictData struct {
	Id           string             `json:"id" bson:"_id"`                      // ID, also UserID
	PredictDate  string             `json:"predict_date" bson:"predict_date"`   // 预测日期 (YYYY-MM-DD)
	PredictPrice float64            `json:"predict_price" bson:"predict_price"` // 预测的价格
	RealPrice    float64            `json:"real_price" bson:"real_price"`       // 真实价格 (开奖时写入)
	IsBot        bool               `json:"-" bson:"is_bot"`                    // 是否机器人
	Status       PricePredictStatus `bson:"status" json:"status"`               // 状态
	BetAmount    int64              `bson:"bet_amount" json:"bet_amount"`       // 下注金额
	RewardPoints int64              `bson:"reward_points" json:"reward_points"` // 奖励积分(预计/实际)
	CreateTime   int64              `bson:"create_time" json:"create_time"`     // 创建时间
	UpdateTime   int64              `json:"update_time" bson:"update_time"`     // 更新时间
}

// GenPricePredictId generates an ID strictly following the rule: pp_userid_predictdate
func GenPricePredictId(userId, predictDate string) string {
	return fmt.Sprintf("pp_%s_%s", userId, predictDate)
}

func NewPricePredictData(userId, predictDate string, predictPrice float64) *PricePredictData {
	now := time.Now().UTC()

	return &PricePredictData{
		Id:           userId, // Id is now UserId
		PredictDate:  predictDate,
		PredictPrice: predictPrice,
		RealPrice:    0,
		Status:       PricePredictStatusUnsubmitted,
		RewardPoints: 0,
		CreateTime:   now.Unix(),
		UpdateTime:   now.Unix(),
	}
}
