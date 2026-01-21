package data

const (
	CollectionExchangeLog = "exchange_log"

	ExchangeStatusUnissued = 0 // 未发放
	ExchangeStatusIssued   = 1 // 发放成功
	ExchangeStatusFailed   = 2 // 发放失败
)

type ExchangeLog struct {
	Id          string  `bson:"_id,omitempty" json:"id"`
	UserId      string  `bson:"user_id" json:"user_id"`
	Date        string  `bson:"date" json:"date"`                 // YYYY-MM-DD
	Amount      float64 `bson:"amount" json:"amount"`             // Points consumed
	TokenAmount float64 `bson:"token_amount" json:"token_amount"` // Tokens received
	WalletAddr  string  `bson:"wallet_addr" json:"wallet_addr"`   // Wallet Address for token distribution
	Status      int     `bson:"status" json:"status"`             // 0: Unissued, 1: Issued, 2: Failed
	TxHash      string  `bson:"tx_hash" json:"tx_hash"`           // Transaction Hash
	CreateTime  int64   `bson:"create_time" json:"create_time"`
}
