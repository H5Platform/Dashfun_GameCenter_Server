package data

type AirdropData struct {
	UserId        string `bson:"_id"`            // 用户ID
	KuCoinId      string `bson:"ku_coin_id"`     // KuCoin ID，用户自行绑定
	TokenAmount   string `bson:"token_amount"`   // airdrop的代币数量，单位为ether
	WalletAddress string `bson:"wallet_address"` // 钱包地址
	ClaimedTime   int64  `bson:"claimed_time"`   // 领取时间，单位为秒，0=未领取
	ClaimTxHash   string `bson:"claim_tx_hash"`  // 领取的交易hash
}
