package data

type SpinWheelRewardType int
type SpinWheelUserStatus int

const (
	SpinWheelReward_GamePoint SpinWheelRewardType = iota + 1 //奖励游戏绑定的积分
)

const (
	SpinWheelUserStatus_Spin     SpinWheelUserStatus = iota + 1 //可转
	SpinWheelUserStatus_CanClaim                                //可以领取奖励
	SpinWheelUserStatus_Claimed                                 //已领取
)

type SpinWheelReward struct {
	RewardIndex int                 `json:"reward_index" bson:"reward_index"` //区域Id, 0~9
	RewardType  SpinWheelRewardType `json:"reward_type" bson:"reward_type"`
	RewardValue int                 `json:"reward_value" bson:"reward_value"`
	Weight      int                 `json:"-" bson:"weight"`
}

// SpinWheelData 轮盘定义，共分10个区域
type SpinWheelData struct {
	Id      string            `json:"id" bson:"_id"`
	Name    string            `json:"name" bson:"name"`
	GameId  string            `json:"game_id" bson:"game_id"` //绑定的游戏Id
	Rewards []SpinWheelReward `json:"rewards" bson:"rewards"` //轮盘每个区域的奖励
}

// SpinWheelUserData 轮盘用户数据
type SpinWheelUserData struct {
	UserId      string              `json:"user_id" bson:"user_id"`
	SpinWheelId string              `json:"spin_wheel_id" bson:"spin_wheel_id"`
	RewardIndex int                 `json:"reward_index" bson:"reward_index"` //抽中的区域索引，0-9
	SpinTime    int64               `json:"spin_time" bson:"spin_time"`       //抽奖时间
	Status      SpinWheelUserStatus `json:"status" bson:"status"`             //状态
}
