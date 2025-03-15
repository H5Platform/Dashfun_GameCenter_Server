package data

type RechargeFrom int
type RechargeStatus int

const (
	DashFunRechargeFrom_TEST RechargeFrom = iota
	DashFunRechargeFrom_TG
)

const (
	DashFunRechargeStatus_Created   RechargeStatus = iota + 1 //订单已创建
	DashFunRechargeStatus_Pending                             //等待支付
	DashFunRechargeStatus_Failed                              //订单失败
	DashFunRechargeStatus_Canceled                            //订单已取消
	DashFunRechargeStatus_Paid                                //订单已支付
	DashFunRechargeStatus_Completed                           //订单已完成，钻石已发放
)

type DashFunRechargeData struct {
	Id           string         `json:"id" bson:"_id"`                //订单ID
	UserId       string         `json:"user_id" bson:"userId"`        //请求订单的用户Id
	PayFrom      string         `json:"pay_from" bson:"payFrom"`      //支付来源
	ChannelPayId string         `json:"payment_id" bson:"paymentId"`  //渠道方的支付Id
	From         RechargeFrom   `json:"from" bson:"from"`             //付款来源渠道
	Price        int            `json:"price" bson:"price"`           //付款金额，单位美分
	Diamond      int            `json:"diamond" bson:"diamond"`       //获得的钻石数量
	Payload      string         `json:"payload" bson:"payload"`       //携带的自定义数据
	Message      string         `json:"message" bson:"message"`       //支付信息，例如失败信息等
	CreatedAt    int64          `json:"created_at" bson:"created_at"` //订单创建时间
	PaidAt       int64          `json:"pay_at" bson:"pay_at"`         //订单支付时间(或者取消时间)
	Status       RechargeStatus `json:"status" bson:"status"`         //订单状态
}
