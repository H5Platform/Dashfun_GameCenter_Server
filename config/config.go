package config

import (
	"dashfun_gamecenter/datasource/data"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

type BaseConfig struct {
	Env string `yaml:"env"` // Dev or Prod
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

type TencentCosConfig struct {
	BucketUrl  string `yaml:"bucket_url"`
	ServiceUrl string `yaml:"service_url"`
	SecretId   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
}

type TonConfig struct {
	ApiKey         string `yaml:"api_key"`
	WalletMnemonic string `yaml:"wallet_mnemonic"`
	WalletVersion  string `yaml:"wallet_version"`
	IsTest         bool   `yaml:"is_test"`
}

type RedisCfg struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type NacosCfg struct {
	IpAddr string `yaml:"ip_addr"`
	Port   uint64 `yaml:"port"`
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

type RechargeOption struct {
	Price        int `yaml:"price"`         //价格，单位美分，浏览器统一价格
	PriceIos     int `yaml:"price_ios"`     //价格，单位美分，AppStore价格
	PriceAndroid int `yaml:"price_android"` //价格，单位美分，AndroidStore价格
	PriceOff     int `yaml:"price_off"`     //折扣(100 = 10% off, 500 = 50% off)
	TGStar       int `yaml:"tg_star"`       //对应TG星星数量
	Diamond      int `yaml:"diamond"`       //对应钻石数量
}

type RechargeCfg struct {
	Open       bool             `yaml:"open"`        //是否开启充值
	EnableStar bool             `yaml:"enable_star"` //是否开启星星充值
	Options    []RechargeOption `yaml:"options"`     //充值选项
}
type StripeConfig struct {
	PublicKey  string `yaml:"public_key"`
	SecretKey  string `yaml:"secret_key"`
	ReturnHost string `yaml:"return_host"`
	WebhookKey string `yaml:"webhook_key"`
}

// CoinCfg 代币配置,DashFunXp, DashFunCoin, DashFunDiamond,DashFunTicket 这些必须代币的配置
type CoinCfg struct {
	Name   string `yaml:"name"`   //代币名称
	Desc   string `yaml:"desc"`   //代币描述
	Symbol string `yaml:"symbol"` //代币符号
}

type SpinWheelCfg struct {
	TicketPrice   int                    `yaml:"ticket_price"`                          //每张票的价格，DashFunDiamond
	TicketsNeeded []int                  `yaml:"tickets_needed"`                        //每次抽奖需要的票数，能抽的次数就是数组的长度，次数每日重置
	Rewards       []data.SpinWheelReward `json:"rewards" bson:"rewards" yaml:"rewards"` //轮盘每个区域的奖励
}
type LeaderboardBotCfg struct {
	RecordScoreMin int                    `yaml:"record_score_min"` //上榜分数的最小值
	BotLevels      []*LeaderboardBotLevel `yaml:"bot_levels"`       //等级配置
}

type LeaderboardBotLevel struct {
	Level               int   `yaml:"level"`                  //等级
	Weight              int   `yaml:"weight"`                 //权重
	MinScore            int   `yaml:"min_score"`              //初始化最小分数
	FixedTaskTop        int   `yaml:"fixed_task_top"`         //固定任务分数上限
	SpinWheelDailyCount int   `yaml:"spin_wheel_daily_count"` //每日转盘次数
	DailyTop            []int `yaml:"daily_top"`              //日常任务分数上限，根据激活天数递减，最后一个数据作为每天的分数
}

type LeaderboardCfg struct {
	Name       string                     `yaml:"name"`        //排行榜名称
	ScoreType  data.LeaderboardScoreType  `yaml:"score_type"`  //排行榜分数类型，由使用者定义，当上报分数给Leaderboard时，会更新所有匹配这个ScoreType的排行榜
	PeriodType data.LeaderboardPeriodType `yaml:"period_type"` //排行榜周期类型
	GameId     string                     `yaml:"game_id"`     //绑定的游戏ID，空或者DashFun表示DashFun平台
	TopCount   int                        `yaml:"top_count"`   //排行榜显示前多少名
}

type PayPalApiBase string

const (
	PayPalApiBaseLive    PayPalApiBase = "live"
	PayPalApiBaseSandbox PayPalApiBase = "sandbox"
)

type PaypalConfig struct {
	ApiBase   PayPalApiBase `yaml:"api_base"`
	ClientId  string        `yaml:"client_id"`
	SecretKey string        `yaml:"secret_key"`
}
type Config struct {
	Base              *BaseConfig        `yaml:"base"`
	Mongo             *MongoConfig       `yaml:"mongo"`
	Web               *WebConfig         `yaml:"web"`
	TG                *TelegramConfig    `yaml:"telegram"`
	Log               *Log               `yaml:"log"`
	PinPoint          *AwsPinPoint       `yaml:"aws_pinpoint"`
	AdminCfg          *AdminConfig       `yaml:"admin_cfg"`
	TencentCosCfg     *TencentCosConfig  `yaml:"tencent_cos"`
	TonCfg            *TonConfig         `yaml:"ton_cfg"`
	RedisCfg          *RedisCfg          `yaml:"redis_cfg"`
	InviteCfg         *InviteCfg         `yaml:"invite_cfg"`
	RechargeCfg       *RechargeCfg       `yaml:"recharge_cfg"`
	CoinCfg           []*CoinCfg         `yaml:"coin_cfg"`
	SpinWheelCfg      *SpinWheelCfg      `yaml:"spin_wheel_cfg"`
	LeaderboardCfg    []*LeaderboardCfg  `yaml:"leaderboard_cfg"` //排行榜配置
	LeaderboardBotCfg *LeaderboardBotCfg `yaml:"leaderboard_bot_cfg"`
	StripeCfg         *StripeConfig      `yaml:"stripe_cfg"`
	NacosCfg          *NacosCfg          `yaml:"nacos_cfg"`
	PaypalCfg         *PaypalConfig      `yaml:"paypal_cfg"`
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

func NacosNamespace() string {
	return strings.ToLower(GetConfig().Base.Env)
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
