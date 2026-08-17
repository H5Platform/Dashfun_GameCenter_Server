package config

import (
	"dashfun_gamecenter/datasource/data"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Mode string

const (
	ModeDashFun   Mode = "DashFun"
	ModeFishVerse Mode = "FishVerse"
	ModeNolanDev  Mode = "NolanDev"
	ModeHowardAI  Mode = "HowardAI"
)

type BaseConfig struct {
	Env  string `yaml:"env"`  // Dev or Prod
	Mode Mode   `yaml:"mode"` // DashFun or FishVerse or NolanDev
}

type WebConfig struct {
	Port int `yaml:"port"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
}

type MongoConfig struct {
	Source   string `yaml:"source"`
	DataBase string `yaml:"data_base"`
}

type Log struct {
	Path string `yaml:"path"`
}

type AwsPinPoint struct {
	KeyId  string `yaml:"key_id"`
	Secret string `yaml:"secret"`
}

type AdminConfig struct {
	Name            string `yaml:"name"`
	Password        string `yaml:"password"`
	BackendPassword string `yaml:"backend_password"`
	ActiveUrl       string `yaml:"active_url"` //账户激活链接
}

type AccountConfig struct {
	TokenSecret string `yaml:"token_secret"`
}

type TencentCosConfig struct {
	BucketUrl  string `yaml:"bucket_url"`
	ServiceUrl string `yaml:"service_url"`
	SecretId   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
}

type RedisCfg struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RewardPoint struct {
	InviteUserType int `yaml:"invite_user_type"` //邀请用户类型, 1=新用户，2=老用户,3=非活跃用户(登录时间超过90天)
	RewardPoint    int `yaml:"reward_point"`     //奖励的积分xp
	RewardCoin     int `yaml:"reward_coin"`      //奖励的金币
}

// InviteCfg 邀请用户奖励配置
type InviteCfg struct {
	PointRequired int           `yaml:"point_required"` //被邀请的用户，达到指定积分后，才算邀请成功
	PointReward   []RewardPoint `yaml:"point_reward"`   //每个成功被邀请的用户奖励的分数
}

// CoinCfg 代币配置,DashFunXp, DashFunCoin, DashFunDiamond,DashFunTicket 这些必须代币的配置
type CoinCfg struct {
	Name        string  `yaml:"name"`         //代币名称
	Desc        string  `yaml:"desc"`         //代币描述
	Symbol      string  `yaml:"symbol"`       //代币符号
	CanWithdraw bool    `yaml:"can_withdraw"` //是否允许提现
	MinWithdraw float32 `yaml:"min_withdraw"` //最低提现数量
}

type LeaderboardBotCfg struct {
	RecordScoreMin int                    `yaml:"record_score_min"` //上榜分数的最小值
	BotLevels      []*LeaderboardBotLevel `yaml:"bot_levels"`       //等级配置
}

type LeaderboardBotLevel struct {
	Level        int   `yaml:"level"`          //等级
	Weight       int   `yaml:"weight"`         //权重
	MinScore     int   `yaml:"min_score"`      //初始化最小分数
	FixedTaskTop int   `yaml:"fixed_task_top"` //固定任务分数上限
	DailyTop     []int `yaml:"daily_top"`      //日常任务分数上限，根据激活天数递减，最后一个数据作为每天的分数
}

type LeaderboardCfg struct {
	Name       string                     `yaml:"name"`        //排行榜名称
	ScoreType  data.LeaderboardScoreType  `yaml:"score_type"`  //排行榜分数类型，由使用者定义，当上报分数给Leaderboard时，会更新所有匹配这个ScoreType的排行榜
	PeriodType data.LeaderboardPeriodType `yaml:"period_type"` //排行榜周期类型
	GameId     string                     `yaml:"game_id"`     //绑定的游戏ID，空或者DashFun表示DashFun平台
	TopCount   int                        `yaml:"top_count"`   //排行榜显示前多少名
}

type Web3Config struct {
	RpcUrl     string `yaml:"rpc_url"`
	ChainId    int    `yaml:"chain_id"`
	PrivateKey string `yaml:"private_key"`
}

type OpenApiConfig struct {
	ApiKey string `yaml:"api_key"`
}

type CoinGeckoConfig struct {
	DefaultTokenIds []string `yaml:"default_token_ids"` // 默认关注的token id列表
	UpdateInterval  int      `yaml:"update_interval"`   // 更新间隔，单位分钟
}

type ForecastConfig struct {
	Url string `yaml:"url"`
}

type PricePredictConfig struct {
	Open            bool    `yaml:"open" json:"open"`                         // 是否开启
	Name            string  `yaml:"name" json:"name"`                         // 预测名称
	Symbol          string  `yaml:"symbol" json:"symbol"`                     // 预测币种符号
	BetStartTime    int     `yaml:"bet_start_time" json:"bet_start_time"`     // 下注开始时间(小时)
	BetEndTime      int     `yaml:"bet_end_time" json:"bet_end_time"`         // 下注结束时间(小时)
	RevealTime      int     `yaml:"reveal_time" json:"reveal_time"`           // 开奖时间(小时)
	MaxDiffLimit    float64 `yaml:"max_diff_limit" json:"max_diff_limit"`     // 最大误差限制(%)
	BetAmounts      []int64 `yaml:"bet_amounts" json:"bet_amounts"`           // 可选下注金额列表
	ConsolationRate float64 `yaml:"consolation_rate" json:"consolation_rate"` // 阳光普照奖比例，下注额度的百分比
	MaxBotCount     int     `yaml:"max_bot_count"`                            // 机器人最大数量
}

type PointExchangeConfig struct {
	PointName        string  `yaml:"point_name" json:"point_name"`                 // 消耗积分名称
	TokenName        string  `yaml:"token_name" json:"token_name"`                 // 兑换Token名称
	TokenAddress     string  `yaml:"token_address" json:"token_address"`           // Token合约地址
	StartTime        string  `yaml:"start_time" json:"start_time"`                 // 开始时间 YYYY-MM-DD-HH
	DurationDays     int     `yaml:"duration_days" json:"duration_days"`           // 持续天数
	DailyGlobalLimit float64 `yaml:"daily_global_limit" json:"daily_global_limit"` // 每日全网限额
	DailyUserLimit   float64 `yaml:"daily_user_limit" json:"daily_user_limit"`     // 每日单人限额
	ExchangeRate     float64 `yaml:"exchange_rate" json:"exchange_rate"`           // 兑换比例 (1 Token = X Point)
}

type HourlySquadGameConfig struct {
	Open             bool   `yaml:"open" json:"open"`                             // 是否开启Bot自动参与
	MinIntervalHours int    `yaml:"min_interval_hours" json:"min_interval_hours"` // Bot参与游戏的最小间隔小时数。平均20小时对应约250有效日活
	MaxIntervalHours int    `yaml:"max_interval_hours" json:"max_interval_hours"` // Bot参与游戏的最大间隔小时数
	ContractAddress  string `yaml:"contract_address" json:"contract_address"`
	AdminPrivateKey  string `yaml:"admin_private_key" json:"admin_private_key"`
}

// GetStartTimeUnix converts StartTime string to Unix timestamp
func (c *PointExchangeConfig) GetStartTimeUnix() int64 {
	// Parse YYYY-MM-DD-HH (Local/Configured Timezone? Usually UTC in this system)
	// System uses UTC generally.
	t, err := time.Parse("2006-01-02-15", c.StartTime)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func (c *PointExchangeConfig) IsActive() bool {
	startUnix := c.GetStartTimeUnix()
	if startUnix == 0 {
		return false
	}
	startTime := time.Unix(startUnix, 0).UTC()
	endTime := startTime.AddDate(0, 0, c.DurationDays)
	now := time.Now().UTC()
	return now.After(startTime) && now.Before(endTime)
}

type Config struct {
	Base                *BaseConfig            `yaml:"base"`
	Mongo               *MongoConfig           `yaml:"mongo"`
	Web                 *WebConfig             `yaml:"web"`
	TG                  *TelegramConfig        `yaml:"telegram"`
	Log                 *Log                   `yaml:"log"`
	PinPoint            *AwsPinPoint           `yaml:"aws_pinpoint"`
	AdminCfg            *AdminConfig           `yaml:"admin_cfg"`
	AccountCfg          *AccountConfig         `yaml:"account_cfg"`
	TencentCosCfg       *TencentCosConfig      `yaml:"tencent_cos"`
	RedisCfg            *RedisCfg              `yaml:"redis_cfg"`
	InviteCfg           *InviteCfg             `yaml:"invite_cfg"`
	CoinCfg             []*CoinCfg             `yaml:"coin_cfg"`
	LeaderboardCfg      []*LeaderboardCfg      `yaml:"leaderboard_cfg"` //排行榜配置
	LeaderboardBotCfg   *LeaderboardBotCfg     `yaml:"leaderboard_bot_cfg"`
	Web3Config          *Web3Config            `yaml:"web3_cfg"`
	OpenApiConfig       *OpenApiConfig         `yaml:"open_api_cfg"`
	CoinGeckoConfig     *CoinGeckoConfig       `yaml:"coingecko_cfg"`
	PricePredictConfig  *PricePredictConfig    `yaml:"price_predict_cfg"`
	PointExchangeConfig *PointExchangeConfig   `yaml:"point_exchange_cfg"`
	ForecastConfig      *ForecastConfig        `yaml:"forecast_cfg"`
	HourlySquadGameCfg  *HourlySquadGameConfig `yaml:"hourly_squad_game_cfg"`
}

var config *Config
var secrets map[string]string

func GetConfig() *Config {
	return config
}

func IsProd() bool {
	return GetConfig().Base.Env == "Prod"
}

func IsTest() bool {
	return GetConfig().Base.Env == "Test"
}

func IsDev() bool {
	return GetConfig().Base.Env == "Dev"
}

func RunningMode() Mode {
	mode := GetConfig().Base.Mode
	if mode == "" {
		return ModeDashFun
	}
	return mode
}

func init() {
	if config == nil {
		load()
	}
}

func load() {
	secrets = make(map[string]string)
	bytes, err := os.ReadFile("./conf/secret.yml")
	if err != nil {
		log.Printf("secret.yml not found: %v\n", err)
	} else {
		yerr := yaml.Unmarshal(bytes, &secrets)
		if yerr != nil {
			log.Fatalf("unmarshal secrets file error: %v\n", err)
		}
	}

	cfg := &Config{}
	bytes, err = os.ReadFile("./conf/config.yml")
	if err != nil {
		log.Fatalf("load config file failed: %v\n", err)
	}

	cfgStr := string(bytes)

	for key, secret := range secrets {
		r := "${" + key + "}"
		cfgStr = strings.ReplaceAll(cfgStr, r, secret)
	}

	bytes = []byte(cfgStr)

	yerr := yaml.Unmarshal(bytes, cfg)
	if yerr != nil {
		log.Fatalf("unmarshal config file failed: %v\n", yerr)
		return
	}

	config = cfg
}
