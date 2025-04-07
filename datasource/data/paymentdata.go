package data

type PaymentFrom int
type PaymentStatus int
type PaymentCurrency string

const (
	DashFunPaymentFrom_TEST PaymentFrom = iota
	DashFunPaymentFrom_DashFun
	DashFunPaymentFrom_TG
)

const (
	DashFunPaymentStatus_Created  PaymentStatus = iota + 1 //订单已创建
	DashFunPaymentStatus_Pending                           //等待支付
	DashFunPaymentStatus_Paid                              //订单已支付
	DashFunPaymentStatus_Canceled                          //订单已取消
	DashFunPaymentStatus_Failed                            //订单失败
)

const (
	PaymentCurrency_DFD          PaymentCurrency = "DFD" //DashFunDiamond
	PaymentCurrency_DFD_TEST     PaymentCurrency = "DFD_TEST"
	PaymentCurrency_TG_STAR      PaymentCurrency = "TG_XTR"      //tg星星
	PaymentCurrency_TG_STAR_TEST PaymentCurrency = "TG_XTR_TEST" //tg金币
)

type DashFunPaymentData struct {
	Id          string          `json:"id" bson:"_id"`                  //订单ID
	UserId      string          `json:"userId" bson:"userId"`           //请求订单的用户Id
	GameId      string          `json:"game_id" bson:"game_id"`         //订单所属GameId
	PaymentId   string          `json:"payment_id" bson:"paymentId"`    //渠道方的支付Id
	Title       string          `json:"title" bson:"title"`             //商品名称
	Description string          `json:"description" bson:"description"` //商品介绍
	Payload     string          `json:"payload" bson:"payload"`         //携带的自定义数据
	Currency    PaymentCurrency `json:"currency" bson:"currency"`       //付款货币
	From        PaymentFrom     `json:"from" bson:"from"`               //付款来源渠道
	Price       int             `json:"price" bson:"price"`             //付款金额
	ExtraData   string          `json:"extraData" bson:"extraData"`     //附加信息
	Message     string          `json:"message" bson:"message"`         //支付信息，例如失败信息等
	CreatedAt   int64           `json:"created_at" bson:"created_at"`   //订单创建时间
	PaidAt      int64           `json:"pay_at" bson:"pay_at"`           //订单支付时间
	Status      PaymentStatus   `json:"status" bson:"status"`           //订单状态
}
