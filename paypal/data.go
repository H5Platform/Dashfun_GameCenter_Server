package paypal

import (
	"strconv"
	"time"
)

type TokenData struct {
	token      string
	createTime time.Time
	expireTime time.Time
}

func NewToken(token string, expiresIn int) Token {
	t := &TokenData{token: token, createTime: time.Now()}
	t.expireTime = t.createTime.Add(time.Duration(expiresIn) * time.Second)
	return t
}

func (t *TokenData) IsExpired() bool {
	return t.expireTime.Before(time.Now())
}

func (t *TokenData) Token() string {
	return t.token
}

type Token interface {
	Token() string
	IsExpired() bool
}

type PurchaseUnitAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type PurchaseUnit struct {
	Amount *PurchaseUnitAmount `json:"amount"`
}

type ExperienceContext struct {
	PaymentMethodPreference string `json:"payment_method_preference"`
	BrandName               string `json:"brand_name"`
	Locale                  string `json:"locale"`
	LandingPage             string `json:"landing_page"`
	ShippingPreference      string `json:"shipping_preference"`
	UserAction              string `json:"user_action"`
}

type OrderRequest struct {
	Intent        string                        `json:"intent"`
	PurchaseUnits []PurchaseUnit                `json:"purchase_units"`
	PaymentSource map[string]*ExperienceContext `json:"payment_source"`
}

func NewOrderRequest(intent string, productName string, price float64) *OrderRequest {
	req := &OrderRequest{
		Intent:        intent,
		PurchaseUnits: make([]PurchaseUnit, 0),
		PaymentSource: make(map[string]*ExperienceContext),
	}
	req.AddPaypalPayment(productName)
	req.AddPurchaseUnit("USD", strconv.FormatFloat(price, 'f', 2, 32))
	return req
}

func (o *OrderRequest) AddPaymentSource(sourceName string, productName string) {
	o.PaymentSource[sourceName] = makePaymentExperience(productName)
}

func (o *OrderRequest) AddPaypalPayment(productName string) {
	o.AddPaymentSource("paypal", productName)
}

func (o *OrderRequest) AddPurchaseUnit(currencyCode string, value string) {
	o.PurchaseUnits = append(o.PurchaseUnits, PurchaseUnit{Amount: &PurchaseUnitAmount{CurrencyCode: currencyCode, Value: value}})
}

func makePaymentExperience(productName string) *ExperienceContext {
	cnt := &ExperienceContext{
		PaymentMethodPreference: "IMMEDIATE_PAYMENT_REQUIRED",
		Locale:                  "en-US",
		LandingPage:             "LOGIN",
		ShippingPreference:      "NO_SHIPPING",
		UserAction:              "PAY_NOW",
		BrandName:               productName,
	}
	return cnt
}
