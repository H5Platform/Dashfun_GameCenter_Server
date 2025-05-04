package paypal

import (
	"bytes"
	"dashfun_gamecenter/config"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	apiBase ApiBase
	token   Token
}

func NewClient(clientId, secret string, apiBase config.PayPalApiBase) *Client {
	c := &Client{
		apiBase: GetApi(apiBase, clientId, secret),
	}
	return c
}

func (c *Client) requestToken() (Token, error) {
	if c.apiBase == nil {
		return nil, errors.New("api base is nil")
	}

	apiUrl := c.apiBase.OauthTokenUrl()

	params := url.Values{}
	params.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	authStr := c.apiBase.ClientId() + ":" + c.apiBase.SecretKey()

	basicAuth := base64.URLEncoding.EncodeToString([]byte(authStr))

	req.Header.Set("Authorization", "Basic "+basicAuth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	jsonStr := string(body)

	accessToken := gjson.Get(jsonStr, "access_token").String()
	expiresIn := int(gjson.Get(jsonStr, "expires_in").Int())

	token := NewToken(accessToken, expiresIn)

	return token, nil
}

func (c *Client) getToken() Token {
	var err error
	if c.token == nil || c.token.IsExpired() {
		c.token, err = c.requestToken()
		if err != nil {
			zap.S().Error("request token failed : " + err.Error())
			return nil
		}
	}
	return c.token
}

func (c *Client) RequestOrder(productName string, price float64) (orderId string, err error) {
	t := c.getToken()
	apiUrl := c.apiBase.RequestOrderUrl()

	orderReq := NewOrderRequest("CAPTURE", productName, price)

	b, _ := json.Marshal(orderReq)

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(b))
	if err != nil {
		zap.S().Error("Paypal requestOrder failed:" + err.Error())
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.Token())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Error("Paypal requestOrder failed:" + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	jsonStr := string(body)

	orderId = gjson.Get(jsonStr, "id").String()
	return orderId, nil
}

func (c *Client) OrderDetails(paypalOrder string) (string, error) {
	t := c.getToken()
	apiUrl := c.apiBase.OrderDetailUrl(paypalOrder)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		zap.S().Error("Paypal OrderDetails [" + paypalOrder + "] failed:" + err.Error())
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.Token())

	zap.S().Infow("Paypal OrderDetails", "orderId", paypalOrder, "token", t.Token())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Error("Paypal OrderDetails [" + paypalOrder + "] failed:" + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	jsonStr := string(body)

	return jsonStr, nil
}

func (c *Client) ConfirmOrder(orderId string) (string, error) {
	t := c.getToken()
	apiUrl := c.apiBase.CaptureOrderUrl(orderId)

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader("{}"))
	if err != nil {
		zap.S().Error("Paypal ConfirmOrder [" + orderId + "] failed:" + err.Error())
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.Token())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Error("Paypal ConfirmOrder [" + orderId + "] failed:" + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	jsonStr := string(body)

	status := gjson.Get(jsonStr, "status").String()
	return status, nil
	//if strings.ToUpper(status) == "COMPLETED" {
	//	op, err := orderService.ConfirmedOrder(orderId)
	//	if err != nil {
	//		log.Println("ConfirmOrder [" + orderId + "] failed:" + err.Error())
	//		return nil, err
	//	}
	//	return op, nil
	//} else {
	//	log.Println("ConfirmOrder [" + orderId + "] failed Status:" + resp.Status)
	//	return nil, errors.New("ConfirmOrder [" + orderId + "] failed Status:" + resp.Status)
	//}
}
