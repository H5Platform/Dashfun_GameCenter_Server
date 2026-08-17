package accountcenter

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"dashfun_gamecenter/apperrors"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/datasource/mongoimpl"
	"dashfun_gamecenter/pinpoint"
	"dashfun_gamecenter/snowflake"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const accountTokenTTL = 30 * 24 * time.Hour
const verifyCodeTTL = 5 * time.Minute

type AccountCenter struct {
	accounts *mongo.Collection
	idGen    *snowflake.Worker
}

var once sync.Once
var inst *AccountCenter

func Get() *AccountCenter {
	once.Do(func() {
		db := mongoimpl.GetMongoDatabase()
		inst = &AccountCenter{accounts: db.Collection("account_data"), idGen: snowflake.Must(snowflake.GetWorker(data.WorkerAccountId))}
		inst.ensureIndexes()
	})
	return inst
}

func (a *AccountCenter) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = a.accounts.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "type", Value: 1}, {Key: "normalized_username", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("type_username_unique")})
	_, _ = a.accounts.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "client_id", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("client_id_unique")})
}

func normalizeUsername(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

func (a *AccountCenter) newAccountId() string {
	return "AC" + strconv.FormatInt(a.idGen.NextId(), 36)
}

func (a *AccountCenter) CreateAccount(username, password string, typ DashFunAccountType) (*AccountResult, error) {
	if typ != DashFunAccountTypeEmail {
		return nil, errors.New("only email account creation is supported")
	}
	if normalizeUsername(username) == "" || len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	accountID := a.newAccountId()
	acc := &accountData{ID: accountID, ClientID: accountID, Username: strings.TrimSpace(username), NormalizedUsername: normalizeUsername(username), PasswordHash: string(passwordHash), Type: typ, Status: DashFunAccountStatusUnvalidated, DisplayName: strings.TrimSpace(username), CreatedAt: now, UpdatedAt: now}
	_, err = a.accounts.InsertOne(context.Background(), acc)
	if mongo.IsDuplicateKeyError(err) {
		return nil, apperrors.ErrAccountAlreadyExists
	}
	if err != nil {
		return nil, err
	}
	return accountResult(acc, ""), nil
}

func (a *AccountCenter) LoginAccount(username, password string, typ DashFunAccountType) (*AccountResult, error) {
	acc, err := a.findByUsername(username, typ)
	if err != nil {
		return nil, err
	}
	if err = a.verifyAndUpgradePassword(acc, password); err != nil {
		return nil, err
	}
	if acc.Status == DashFunAccountStatusUnvalidated {
		return accountResult(acc, ""), apperrors.ErrAccountUnvalidated
	}
	if acc.Status == DashFunAccountStatusFrozen {
		return nil, apperrors.ErrAccountFrozen
	}
	if acc.Status != DashFunAccountStatusNormal {
		return nil, apperrors.ErrAccountNotFound
	}
	return accountResult(acc, a.issueToken(acc)), nil
}

func (a *AccountCenter) RequestSendVerifyEmail(accountID string) error {
	acc, err := a.findByID(accountID)
	if err != nil {
		return err
	}
	code, err := randomCode()
	if err != nil {
		return err
	}
	_, err = a.accounts.UpdateByID(context.Background(), acc.ID, bson.M{"$set": bson.M{"verify_code_hash": hashString(code), "verify_code_expires": time.Now().UTC().Add(verifyCodeTTL), "updated_at": time.Now().UTC()}})
	if err != nil {
		return err
	}
	return pinpoint.Get().SendEmail("DashFun verification code", acc.Username, fmt.Sprintf("Your verification code is %s. It expires in 5 minutes.", code))
}

func (a *AccountCenter) VerifyEmailCode(accountID, code string) (*AccountResult, error) {
	acc, err := a.findByID(accountID)
	if err != nil {
		return nil, err
	}
	if !validCode(acc.VerifyCodeHash, acc.VerifyCodeExpires, code) {
		return nil, errors.New("invalid or expired verification code")
	}
	acc.Status = DashFunAccountStatusNormal
	_, err = a.accounts.UpdateByID(context.Background(), acc.ID, bson.M{"$set": bson.M{"status": acc.Status, "updated_at": time.Now().UTC()}, "$unset": bson.M{"verify_code_hash": "", "verify_code_expires": ""}})
	if err != nil {
		return nil, err
	}
	return accountResult(acc, a.issueToken(acc)), nil
}

func (a *AccountCenter) CheckToken(accountID, token string, typ DashFunAccountType) (*AccountResult, error) {
	acc, err := a.AuthenticateToken(token)
	if err != nil || acc.AccountId != accountID || acc.Type != typ {
		return nil, errors.New("invalid account token")
	}
	dbAcc, err := a.findByID(accountID)
	if err != nil {
		return nil, err
	}
	acc.Token = a.issueToken(dbAcc)
	return acc, nil
}

func (a *AccountCenter) AuthenticateToken(token string) (*AccountResult, error) {
	accountID, tokenType, err := verifyAccountToken(token)
	if err != nil {
		return nil, err
	}
	acc, err := a.findByID(accountID)
	if err != nil || acc.Status != DashFunAccountStatusNormal {
		return nil, errors.New("invalid account token")
	}
	if tokenType != 0 && acc.Type != tokenType {
		return nil, errors.New("invalid account token")
	}
	return accountResult(acc, token), nil
}

func (a *AccountCenter) DeleteAccount(accountID, token string, typ DashFunAccountType) error {
	authenticated, err := a.AuthenticateToken(token)
	if err != nil || authenticated.AccountId != accountID || authenticated.Type != typ {
		return errors.New("invalid account token")
	}
	account, err := a.findByID(accountID)
	if err != nil {
		return err
	}
	_, err = a.accounts.UpdateByID(context.Background(), account.ID, bson.M{"$set": bson.M{"status": DashFunAccountStatusDeleted, "normalized_username": "deleted:" + accountID, "updated_at": time.Now().UTC()}})
	return err
}

func (a *AccountCenter) RequestResetPassword(username string) error {
	acc, err := a.findByUsername(username, DashFunAccountTypeEmail)
	if err != nil {
		return nil
	}
	code, err := randomCode()
	if err != nil {
		return err
	}
	_, err = a.accounts.UpdateByID(context.Background(), acc.ID, bson.M{"$set": bson.M{"reset_code_hash": hashString(code), "reset_code_expires": time.Now().UTC().Add(verifyCodeTTL), "updated_at": time.Now().UTC()}})
	if err != nil {
		return err
	}
	return pinpoint.Get().SendEmail("DashFun password reset", acc.Username, fmt.Sprintf("Your password reset code is %s. It expires in 5 minutes.", code))
}

func (a *AccountCenter) ResetPassword(username, code, password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	acc, err := a.findByUsername(username, DashFunAccountTypeEmail)
	if err != nil {
		return err
	}
	if !validCode(acc.ResetCodeHash, acc.ResetCodeExpires, code) {
		return errors.New("invalid or expired reset code")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.accounts.UpdateByID(context.Background(), acc.ID, bson.M{"$set": bson.M{"password_hash": string(passwordHash), "normalized_username": normalizeUsername(acc.Username), "updated_at": time.Now().UTC()}, "$unset": bson.M{"password": "", "reset_code_hash": "", "reset_code_expires": ""}})
	return err
}

func (a *AccountCenter) findByID(id string) (*accountData, error) {
	var v accountData
	if a.accounts.FindOne(context.Background(), bson.M{"$or": bson.A{bson.M{"_id": id}, bson.M{"client_id": id}}}).Decode(&v) != nil {
		return nil, apperrors.ErrAccountNotFound
	}
	if err := a.ensureClientID(&v); err != nil {
		return nil, err
	}
	return &v, nil
}
func (a *AccountCenter) findByUsername(username string, typ DashFunAccountType) (*accountData, error) {
	var v accountData
	err := a.accounts.FindOne(context.Background(), bson.M{"type": typ, "normalized_username": normalizeUsername(username)}).Decode(&v)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = a.accounts.FindOne(context.Background(), bson.M{"type": typ, "username": strings.TrimSpace(username)}).Decode(&v)
	}
	if err != nil {
		return nil, apperrors.ErrAccountNotFound
	}
	if err := a.ensureClientID(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (a *AccountCenter) ensureClientID(acc *accountData) error {
	if acc.ClientID != "" {
		return nil
	}
	clientID := a.newAccountId()
	result, err := a.accounts.UpdateOne(context.Background(), bson.M{"_id": acc.ID, "client_id": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"client_id": clientID, "updated_at": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if result.ModifiedCount == 1 {
		acc.ClientID = clientID
		return nil
	}
	var current accountData
	if err = a.accounts.FindOne(context.Background(), bson.M{"_id": acc.ID}).Decode(&current); err != nil {
		return err
	}
	acc.ClientID = current.ClientID
	return nil
}

func (a *AccountCenter) verifyAndUpgradePassword(acc *accountData, password string) error {
	if acc.PasswordHash != "" {
		if bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password)) != nil {
			return apperrors.ErrAccountPasswordIncorrect
		}
		return nil
	}
	if acc.LegacyPassword == "" || subtle.ConstantTimeCompare([]byte(acc.LegacyPassword), []byte(password)) != 1 {
		return apperrors.ErrAccountPasswordIncorrect
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.accounts.UpdateByID(context.Background(), acc.ID, bson.M{"$set": bson.M{"password_hash": string(passwordHash), "normalized_username": normalizeUsername(acc.Username), "updated_at": time.Now().UTC()}, "$unset": bson.M{"password": ""}})
	return err
}
func (a *AccountCenter) issueToken(acc *accountData) string {
	values := map[string]string{"accountId": acc.publicID(), "username": acc.Username, "type": strconv.Itoa(int(acc.Type)), "auth_date": strconv.FormatInt(time.Now().Unix(), 10), "display_name": acc.DisplayName}
	query := url.Values{}
	for key, value := range values {
		query.Set(key, value)
	}
	query.Set("hash", signToken(values))
	return query.Encode()
}

type browserAuthUser struct {
	AccountID   string `json:"acc_id"`
	Type        string `json:"type"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func verifyAccountToken(token string) (string, DashFunAccountType, error) {
	values, err := url.ParseQuery(token)
	if err != nil {
		return "", 0, errors.New("invalid account token")
	}
	accountID, username, typeValue, displayName := values.Get("accountId"), values.Get("username"), values.Get("type"), values.Get("display_name")
	if rawUser := values.Get("user"); rawUser != "" {
		var user browserAuthUser
		if json.Unmarshal([]byte(rawUser), &user) != nil {
			return "", 0, errors.New("invalid account token")
		}
		accountID, username, typeValue, displayName = user.AccountID, user.Username, user.Type, user.DisplayName
	}
	authDate := values.Get("auth_date")
	signed := map[string]string{"accountId": accountID, "username": username, "type": typeValue, "auth_date": authDate, "display_name": displayName}
	if accountID == "" || authDate == "" || !hmac.Equal([]byte(signToken(signed)), []byte(values.Get("hash"))) {
		return "", 0, errors.New("invalid account token")
	}
	issuedAt, err := strconv.ParseInt(authDate, 10, 64)
	if err != nil || issuedAt > time.Now().Unix()+300 || time.Now().Unix()-issuedAt > int64(accountTokenTTL.Seconds()) {
		return "", 0, errors.New("expired account token")
	}
	typeNumber, err := strconv.Atoi(typeValue)
	if err != nil {
		return "", 0, errors.New("invalid account token")
	}
	return accountID, DashFunAccountType(typeNumber), nil
}

func signToken(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	secret := ""
	if config.GetConfig().AccountCfg != nil {
		secret = config.GetConfig().AccountCfg.TokenSecret
	}
	if secret == "" || strings.HasPrefix(secret, "${") {
		secret = config.GetConfig().TG.Token
	}
	secretKey := sha256.Sum256([]byte(secret))
	mac := hmac.New(sha256.New, secretKey[:])
	_, _ = mac.Write([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(mac.Sum(nil))
}
func accountResult(acc *accountData, token string) *AccountResult {
	return &AccountResult{AccountId: acc.publicID(), Username: acc.Username, Type: acc.Type, Status: acc.Status, Token: token, DisplayName: acc.DisplayName}
}
func hashString(v string) string { sum := sha256.Sum256([]byte(v)); return hex.EncodeToString(sum[:]) }
func validCode(hash string, expires time.Time, code string) bool {
	return hash != "" && time.Now().UTC().Before(expires) && subtle.ConstantTimeCompare([]byte(hash), []byte(hashString(strings.TrimSpace(code)))) == 1
}
func randomCode() (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range raw {
		raw[i] = chars[int(raw[i])%len(chars)]
	}
	return string(raw), nil
}
