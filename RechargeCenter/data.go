package RechargeCenter

import "dashfun_gamecenter/datasource/data"

type RechargePlatformOption struct {
	Price    int `json:"price"`     //价格，单位美分,tg平台是星星数量
	Diamond  int `json:"diamond"`   //钻石数量
	PriceOff int `json:"price_off"` //折扣(10 = 10% off)
}

type RechargePlatformOptions struct {
	PriceType data.RechargePlatformOptionPriceType `json:"price_type"`
	Options   []RechargePlatformOption             `json:"options"`
}
