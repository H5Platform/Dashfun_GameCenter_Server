package AirdropCenter

type AirdropUserDetail struct {
	StartTime       int64  `json:"start_time"`       // 开始时间，单位为秒
	ClaimTime       int64  `json:"claim_time"`       // TGE后多久可以领取，单位秒
	TokenAmount     string `json:"token_amount"`     // 获得的代币数量，单位为ether
	Claimed         bool   `json:"claimed"`          // 是否已领取(已经调用过合约的createVesting方法)
	VestingContract string `json:"vesting_contract"` // 锁仓合约地址
	TokenContract   string `json:"token_contract"`   // 代币合约地址
	ClaimAddress    string `json:"claim_address"`    // 用户提交的领取地址
	KuCoinId        string `json:"ku_coin_id"`       // KuCoin ID，用户自行绑定
}
