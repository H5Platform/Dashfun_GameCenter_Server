package accountcenter

import "time"

type DashFunAccountStatus int

const (
	DashFunAccountStatusUnvalidated DashFunAccountStatus = iota
	DashFunAccountStatusNormal
	DashFunAccountStatusFrozen
	DashFunAccountStatusDeleted
)

type DashFunAccountType int

const (
	DashFunAccountTypeEmail DashFunAccountType = iota + 1
	DashFunAccountTypeTelegram
	DashFunAccountTypeAppleId
)

type AccountResult struct {
	AccountId   string               `json:"account_id"`
	Username    string               `json:"username"`
	Type        DashFunAccountType   `json:"type"`
	Status      DashFunAccountStatus `json:"status"`
	Token       string               `json:"token"`
	DisplayName string               `json:"display_name"`
}

type accountData struct {
	ID                 string               `bson:"_id"`
	ClientID           string               `bson:"client_id,omitempty"`
	Username           string               `bson:"username"`
	NormalizedUsername string               `bson:"normalized_username"`
	PasswordHash       string               `bson:"password_hash"`
	LegacyPassword     string               `bson:"password,omitempty"`
	Type               DashFunAccountType   `bson:"type"`
	Status             DashFunAccountStatus `bson:"status"`
	DisplayName        string               `bson:"display_name"`
	CreatedAt          time.Time            `bson:"created_at"`
	UpdatedAt          time.Time            `bson:"updated_at"`
	VerifyCodeHash     string               `bson:"verify_code_hash,omitempty"`
	VerifyCodeExpires  time.Time            `bson:"verify_code_expires,omitempty"`
	ResetCodeHash      string               `bson:"reset_code_hash,omitempty"`
	ResetCodeExpires   time.Time            `bson:"reset_code_expires,omitempty"`
}

func (a *accountData) publicID() string {
	if a.ClientID != "" {
		return a.ClientID
	}
	return a.ID
}
