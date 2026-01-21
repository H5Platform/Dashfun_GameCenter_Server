package data

import (
	"math/rand"
	"time"
)

type NolanBotData struct {
	Id                  string    `json:"id" bson:"_id"`                                        // Bot ID
	Name                string    `json:"name" bson:"name"`                                     // Bot Name
	Score               int64     `json:"score" bson:"score"`                                   // Bot Score
	Rank                int64     `json:"rank" bson:"rank"`                                     // Bot Rank
	MaxPostIntervalDays int64     `json:"max_post_interval_days" bson:"max_post_interval_days"` // Maximum post interval in days
	MinPostIntervalDays int64     `json:"min_post_interval_days" bson:"min_post_interval_days"` // Minimum post interval in days
	RegionId            int       `json:"region_id" bson:"region_id"`                           // Region ID FishingRegions
	ActiveTime          int64     `json:"active_time" bson:"active_time"`
	WalletAddr          string    `json:"wallet_addr" bson:"wallet_addr"`   // bot的活跃时间，单位毫秒，到这个时间后活跃发帖，然后随机等待一个天数(MinPostIntervalDays, MaxPostIntervalDays)后再次活跃
	TokenAmount         float64   `json:"token_amount" bson:"token_amount"` // Bot Token Amount
	Status              BotStatus `json:"status" bson:"status"`             // Bot Status
}

func (b *NolanBotData) RandomNextActiveTime() {
	intervalDays := rand.Int63n(b.MaxPostIntervalDays-b.MinPostIntervalDays+1) + b.MinPostIntervalDays
	t := time.Now().Add(time.Duration(intervalDays) * 24 * time.Hour)
	t = t.Truncate(24 * time.Hour)
	b.ActiveTime = t.UnixMilli() + rand.Int63n(24*60*60*1000)
}

type NolanPostData struct {
	PostId     string `bson:"_id" json:"postId"`             // 帖子ID
	UserId     string `bson:"user_id" json:"userId"`         // 用户ID
	PosterName string `bson:"poster_name" json:"posterName"` // 帖子发布者名称
	Content    string `bson:"content" json:"content"`        // 帖子内容
	CreatedAt  int64  `bson:"created_at" json:"createdAt"`   // 帖子创建时间
	Location   string `bson:"location" json:"location"`      // 帖子位置
}
