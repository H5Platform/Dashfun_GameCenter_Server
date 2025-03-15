package RechargeCenter

type RechargePlatformOptionPriceType int

const (
	RechargePlatformOptionPriceTypeUSD RechargePlatformOptionPriceType = iota + 1
	RechargePlatformOptionPriceTypeTGStar
)

type RechargePlatformOption struct {
	Price    int `json:"price"`     //价格，单位美分,tg平台是星星数量
	Diamond  int `json:"diamond"`   //钻石数量
	PriceOff int `json:"price_off"` //折扣(10 = 10% off)
}

type RechargePlatformOptions struct {
	PriceType RechargePlatformOptionPriceType `json:"price_type"`
	Options   []RechargePlatformOption        `json:"options"`
}
