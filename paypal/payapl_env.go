package paypal

import (
	"dashfun_gamecenter/config"
	"fmt"
)

type ApiBase interface {
	apiUrl() string
	OauthTokenUrl() string
	RequestOrderUrl() string
	CaptureOrderUrl(paypalOrderId string) string
	ConfirmOrderUrl(paypalOrderId string) string
	OrderDetailUrl(paypalOrderId string) string
	ClientId() string
	SecretKey() string
}

type ApiBaseSandbox struct {
	clientId  string
	secretKey string
}

func (p *ApiBaseSandbox) apiUrl() string {
	return "https://api-m.sandbox.paypal.com"
}

func (p *ApiBaseSandbox) OauthTokenUrl() string {
	return fmt.Sprintf("%s/v1/oauth2/token", p.apiUrl())
}

func (p *ApiBaseSandbox) RequestOrderUrl() string {
	return fmt.Sprintf("%s/v2/checkout/orders", p.apiUrl())
}

func (p *ApiBaseSandbox) CaptureOrderUrl(paypalOrderId string) string {
	return fmt.Sprintf("%s/v2/checkout/orders/%s/capture", p.apiUrl(), paypalOrderId)
}

func (p *ApiBaseSandbox) ConfirmOrderUrl(paypalOrderId string) string {
	return fmt.Sprintf("%s/v2/checkout/orders/%s/confirm-payment-source", p.apiUrl(), paypalOrderId)
}

func (p *ApiBaseSandbox) OrderDetailUrl(paypalOrderId string) string {
	return fmt.Sprintf("%s/v2/checkout/orders/%s", p.apiUrl(), paypalOrderId)
}

func (p *ApiBaseSandbox) ClientId() string {
	return p.clientId
}

func (p *ApiBaseSandbox) SecretKey() string {
	return p.secretKey
}

type ApiBaseLive struct {
	clientId  string
	secretKey string
}

func (p *ApiBaseLive) apiUrl() string {
	return "https://api-m.paypal.com"
}

func (p *ApiBaseLive) OauthTokenUrl() string {
	return fmt.Sprintf("%s/v1/oauth2/token", p.apiUrl())
}

func (p *ApiBaseLive) RequestOrderUrl() string {
	return fmt.Sprintf("%s/v2/checkout/orders", p.apiUrl())
}

func (p *ApiBaseLive) CaptureOrderUrl(paypalOrderId string) string {
	return fmt.Sprintf("%s/v2/checkout/orders/%s/capture", p.apiUrl(), paypalOrderId)
}

func (p *ApiBaseLive) ConfirmOrderUrl(paypalOrderId string) string {
	return fmt.Sprintf("%s/v2/checkout/orders/%s/confirm-payment-source", p.apiUrl(), paypalOrderId)
}

func (p *ApiBaseLive) OrderDetailUrl(paypalOrderId string) string {
	return fmt.Sprintf("%s/v2/checkout/orders/%s", p.apiUrl(), paypalOrderId)
}

func (p *ApiBaseLive) ClientId() string {
	return p.clientId
}

func (p *ApiBaseLive) SecretKey() string {
	return p.secretKey
}

func GetApi(apiBase config.PayPalApiBase, clientId, clientSecret string) ApiBase {
	switch apiBase {
	case config.PayPalApiBaseLive:
		return &ApiBaseLive{
			clientId:  clientId,
			secretKey: clientSecret,
		}
	default:
		return &ApiBaseSandbox{
			clientId:  clientId,
			secretKey: clientSecret,
		}
	}
}
