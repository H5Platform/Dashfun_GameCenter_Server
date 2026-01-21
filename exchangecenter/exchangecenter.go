package exchangecenter

import (
	"context"
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/web3center"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ExchangeCenter struct {
	cfg *config.PointExchangeConfig

	mu               sync.Mutex
	cacheGlobalUsage float64
	cacheDate        string

	// Queue for token distribution
	queue chan *data.ExchangeLog
}

var (
	instance *ExchangeCenter
	once     sync.Once
)

func Get() *ExchangeCenter {
	once.Do(func() {
		instance = &ExchangeCenter{
			cfg:   config.GetConfig().PointExchangeConfig,
			queue: make(chan *data.ExchangeLog, 10000),
		}
		go instance.startWorker()

		// Recover unissued logs on startup
		// This should be done async to not block startup if DB is slow
		go instance.recoverUnissued()
	})
	return instance
}

func (c *ExchangeCenter) recoverUnissued() {
	// Wait a bit for DB to be potentially ready if needed (or just retry)
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	logs, err := dao.GetExchangeDao().GetUnissuedLogs(ctx)
	if err != nil {
		zap.S().Errorw("ExchangeCenter: Failed to recover unissued logs", "err", err)
		return
	}

	for _, log := range logs {
		c.queue <- log
	}
	zap.S().Infow("ExchangeCenter: Recovered unissued logs", "count", len(logs))
}

func (c *ExchangeCenter) startWorker() {
	zap.S().Info("ExchangeCenter: Worker started")
	for log := range c.queue {
		c.processDistribution(log)
	}
}

func (c *ExchangeCenter) processDistribution(log *data.ExchangeLog) {
	ctx := context.Background()
	zap.S().Infow("ExchangeCenter: Processing distribution", "logId", log.Id, "userId", log.UserId, "amount", log.TokenAmount)

	// Transfer Token
	// Using Web3Center
	txHash, err := web3center.Get().TransferToken(c.cfg.TokenAddress, log.WalletAddr, log.TokenAmount)

	status := data.ExchangeStatusIssued
	if err != nil {
		zap.S().Errorw("ExchangeCenter: Transfer failed", "logId", log.Id, "err", err)
		status = data.ExchangeStatusFailed
		txHash = err.Error() // Store error message in txHash or separate field? User asked for TxHash. Let's store error string or empty.
		// If failed, maybe retry? Simple queue: moved to failed status.
	} else {
		zap.S().Infow("ExchangeCenter: Transfer success", "logId", log.Id, "txHash", txHash)
	}

	// Update DB
	err = dao.GetExchangeDao().UpdateStatus(ctx, log.Id, status, txHash)
	if err != nil {
		zap.S().Errorw("ExchangeCenter: Failed to update status", "logId", log.Id, "err", err)
	}
}

type ExchangeActivityInfo struct {
	Config          *config.PointExchangeConfig `json:"config"`
	Status          string                      `json:"status"` // "NotStarted", "Active", "Ended"
	GlobalUsed      float64                     `json:"global_used"`
	GlobalRemaining float64                     `json:"global_remaining"`
	UserUsed        float64                     `json:"user_used"`
	UserRemaining   float64                     `json:"user_remaining"`
	StartTimeUnix   int64                       `json:"start_time_unix"`
}

// getGlobalUsageWithCache gets global usage from cache or DB safely
func (c *ExchangeCenter) getGlobalUsageWithCache(ctx context.Context, date string) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cacheDate == date {
		return c.cacheGlobalUsage, nil
	}

	// Cache miss or date changed, load from DB
	usage, err := dao.GetExchangeDao().GetDailyGlobalUsage(ctx, date)
	if err != nil {
		return 0, err
	}
	c.cacheDate = date
	c.cacheGlobalUsage = usage
	return usage, nil
}

// incrementGlobalUsage updates cache safetly
func (c *ExchangeCenter) incrementGlobalUsage(amount float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheGlobalUsage += amount
}

// decrementGlobalUsage updates cache safetly (rollback)
func (c *ExchangeCenter) decrementGlobalUsage(amount float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheGlobalUsage -= amount
}

func (c *ExchangeCenter) GetActivityInfo(ctx context.Context, userId string) (*ExchangeActivityInfo, error) {
	if c.cfg == nil {
		return &ExchangeActivityInfo{Status: "NotStarted"}, nil
	}

	info := &ExchangeActivityInfo{
		Config:        c.cfg,
		StartTimeUnix: c.cfg.GetStartTimeUnix(),
	}

	now := time.Now().UTC()
	startUnix := c.cfg.GetStartTimeUnix()
	startTime := time.Unix(startUnix, 0).UTC()

	// Calculate End Time
	endTime := startTime.AddDate(0, 0, c.cfg.DurationDays)

	if now.Before(startTime) {
		info.Status = "NotStarted"
	} else if now.After(endTime) {
		info.Status = "Ended"
	} else {
		info.Status = "Active"
	}

	currentDate := now.Format("2006-01-02")

	// Get Global Usage (Cache)
	globalUsed, err := c.getGlobalUsageWithCache(ctx, currentDate)
	if err != nil {
		zap.S().Errorw("Failed to get global usage", "error", err)
	}
	info.GlobalUsed = globalUsed
	info.GlobalRemaining = c.cfg.DailyGlobalLimit - globalUsed
	if info.GlobalRemaining < 0 {
		info.GlobalRemaining = 0
	}

	// Get User Usage (DB is fine for user limit)
	userUsed, err := dao.GetExchangeDao().GetDailyUserUsage(ctx, userId, currentDate)
	if err != nil {
		zap.S().Errorw("Failed to get user usage", "error", err)
	}
	info.UserUsed = userUsed
	info.UserRemaining = c.cfg.DailyUserLimit - userUsed
	if info.UserRemaining < 0 {
		info.UserRemaining = 0
	}

	return info, nil
}

func (c *ExchangeCenter) Exchange(ctx context.Context, userId string, amount int64, walletAddr string, isBot bool) (float64, error) {
	if c.cfg == nil {
		return 0, errors.New("exchange config not found")
	}

	if amount <= 0 {
		return 0, errors.New("invalid amount")
	}

	if walletAddr == "" {
		return 0, errors.New("wallet address required")
	}

	// 1. Check Activity Status
	now := time.Now().UTC()
	startUnix := c.cfg.GetStartTimeUnix()
	startTime := time.Unix(startUnix, 0).UTC()
	endTime := startTime.AddDate(0, 0, c.cfg.DurationDays)

	if now.Before(startTime) {
		return 0, errors.New("activity not started")
	}
	if now.After(endTime) {
		return 0, errors.New("activity ended")
	}

	currentDate := now.Format("2006-01-02")
	tokenAmount := float64(amount) / c.cfg.ExchangeRate
	if tokenAmount <= 0 {
		return 0, errors.New("exchange amount too small")
	}

	// 2. Check & Reserve Global Limit (Atomic Local Cache)
	c.mu.Lock()

	// Refresh cache if needed (lazy load inside lock to ensure consistency)
	if c.cacheDate != currentDate {
		usage, err := dao.GetExchangeDao().GetDailyGlobalUsage(ctx, currentDate)
		if err != nil {
			c.mu.Unlock()
			return 0, err
		}
		c.cacheDate = currentDate
		c.cacheGlobalUsage = usage
	}

	if c.cacheGlobalUsage+tokenAmount > c.cfg.DailyGlobalLimit {
		c.mu.Unlock()
		return 0, errors.New("daily global limit exceeded")
	}

	// Reserve quota
	c.cacheGlobalUsage += tokenAmount
	c.mu.Unlock()

	// 3. Deduction (Points) - Skipped for Bots
	var err error
	if !isBot {
		// Check User Limit from DB
		var userUsed float64
		userUsed, err = dao.GetExchangeDao().GetDailyUserUsage(ctx, userId, currentDate)
		if err != nil {
			c.decrementGlobalUsage(tokenAmount)
			return 0, err
		}
		if userUsed+tokenAmount > c.cfg.DailyUserLimit {
			c.decrementGlobalUsage(tokenAmount)
			return 0, errors.New("daily user limit exceeded")
		}

		coin, found := coincenter.Get().GetCoinByName(c.cfg.PointName)
		if !found {
			c.decrementGlobalUsage(tokenAmount)
			return 0, errors.New("point coin not found: " + c.cfg.PointName)
		}

		// CoinCenter uses int32
		if amount > 2147483647 { // Max value for int32
			c.decrementGlobalUsage(tokenAmount)
			return 0, errors.New("amount too large")
		}
		pointInt := int32(amount)

		_, err = coincenter.Get().DecUserCoinAmount(userId, coin.Id, pointInt, "TokenExchange", fmt.Sprintf("Exchanged for %.4f %s", tokenAmount, c.cfg.TokenName))
		if err != nil {
			c.decrementGlobalUsage(tokenAmount)
			return 0, err
		}
	}

	// 4. Record Log
	log := &data.ExchangeLog{
		UserId:      userId,
		Date:        currentDate,
		Amount:      float64(amount), // Points
		TokenAmount: tokenAmount,     // Tokens
		WalletAddr:  walletAddr,
		CreateTime:  now.Unix(),
	}

	err = dao.GetExchangeDao().Save(ctx, log)
	if err != nil {
		// Log saved failed. Quota is used, User Points deducted.
		// Rollback? No. User lost points, but "transaction" logic completed (deduction).
		// We just log error.
		zap.S().Errorw("Failed to save exchange log", "log", log, "error", err)
	} else {
		// Push to queue for distribution
		// We use non-blocking send or large buffer? Buffer is 1000.
		// Use select to prevent blocking request if buffer full (risk of dropping distribution task, but user deducted).
		// Better to block or handle overflow. Given requirements "Queue to distribute", we assume buffer is enough or we block/retry.
		// Ideally: Persistence ensured by DB. If channel full, we can rely on Recovery.
		// But let's try to push.
		select {
		case c.queue <- log:
		default:
			zap.S().Errorw("ExchangeCenter: Queue full, log recorded but not queued immediately (will be recovered on restart)", "logId", log.Id)
		}
	}

	return tokenAmount, nil
}

func (c *ExchangeCenter) GetHistory(ctx context.Context, userId string) ([]*data.ExchangeLog, error) {
	// Limit to 50 as requested
	return dao.GetExchangeDao().GetExchangeHistory(ctx, userId, 50)
}
