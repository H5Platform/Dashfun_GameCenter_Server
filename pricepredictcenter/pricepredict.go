package pricepredictcenter

import (
	"dashfun_gamecenter/coincenter"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"
)

var oncePricePredictCenter sync.Once
var instPricePredictCenter *PricePredictCenter

const (
	SystemPoolUserId = "SYSTEM_POOL"
	RolloverDate     = "ROLLOVER"
)

type PricePredictCenter struct {
	cfg            *config.PricePredictConfig
	fetcher        *PriceFetcher
	lastRevealDate string // Cache to avoid repeated checks if done

	// Caching today's pool stats
	cacheMutex     sync.RWMutex
	cacheTodayPool int64
	cacheUserCount int64
	bot            *PricePredictBot // Bot instance
}

func Get() *PricePredictCenter {
	oncePricePredictCenter.Do(func() {
		instPricePredictCenter = &PricePredictCenter{}
		instPricePredictCenter.init()
	})
	return instPricePredictCenter
}

func (p *PricePredictCenter) init() {
	p.cfg = config.GetConfig().PricePredictConfig
	if p.cfg == nil {
		zap.S().Warn("PricePredict config is nil, skipping initialization")
		return
	}

	if !p.cfg.Open {
		zap.S().Info("PricePredict is not open")
		return
	}

	p.fetcher = NewPriceFetcher()
	p.bot = NewPricePredictBot(p) // Init Bot

	// Rebuild Cache
	p.rebuildTodayCache()

	p.logConfigDetails()
	zap.S().Infow("PricePredict Center initialized", "name", p.cfg.Name, "symbol", p.cfg.Symbol, "cached_pool", p.cacheTodayPool, "cached_users", p.cacheUserCount)

	go p.startScheduler()
	p.bot.Start()
}

func (p *PricePredictCenter) rebuildTodayCache() {
	date := p.getPredictDate(time.Now())
	users, err := dao.GetPricePredictDao().FindPendingPredictionsByDate(date)
	if err != nil {
		zap.S().Errorw("Failed to rebuild cache", "error", err)
		return
	}
	var total int64
	for _, u := range users {
		total += u.BetAmount
	}
	p.cacheMutex.Lock()
	p.cacheTodayPool = total
	p.cacheUserCount = int64(len(users))
	p.cacheMutex.Unlock()
}

func (p *PricePredictCenter) startScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.checkAndReveal()
	}
}

func (p *PricePredictCenter) getPredictDate(now time.Time) string {
	if config.IsDev() {
		// Dev: Hourly cycles. Format: YYYY-MM-DD-HH
		return now.UTC().Format("2006-01-02-15")
	}
	// Prod: Daily cycles. Format: YYYY-MM-DD
	return now.UTC().Format("2006-01-02")
}

func (p *PricePredictCenter) checkAndReveal() {
	if p.cfg == nil || !p.cfg.Open {
		return
	}
	now := time.Now().UTC()

	shouldReveal := false
	if config.IsDev() {
		// Dev: Check Minute
		if now.Minute() >= int(p.cfg.RevealTime) {
			shouldReveal = true
		}
	} else {
		// Prod: Check Hour
		if now.Hour() >= int(p.cfg.RevealTime) {
			shouldReveal = true
		}
	}

	if shouldReveal {
		// Try to reveal
		p.RevealResult(now)
	}
}

func (p *PricePredictCenter) logConfigDetails() {
	zap.S().Infow("PricePredict Configuration",
		"bet_start", p.cfg.BetStartTime,
		"bet_end", p.cfg.BetEndTime,
		"max_diff_limit", p.cfg.MaxDiffLimit,
		"consolation_rate", p.cfg.ConsolationRate,
		"is_dev", config.IsDev(),
	)
}

func (p *PricePredictCenter) RevealResult(now time.Time) {
	date := p.getPredictDate(now)

	// Optimization: If already done for today, skip
	if p.lastRevealDate == date {
		return
	}

	// 1. Fetch pending predictions
	users, err := dao.GetPricePredictDao().FindPendingPredictionsByDate(date)
	if err != nil {
		zap.S().Errorw("Failed to fetch pending predictions", "error", err)
		return
	}
	if len(users) == 0 {
		// No pending users, mark as done for today
		p.lastRevealDate = date
		return
	}

	zap.S().Infow("Start revealing price prediction", "date", date, "user_count", len(users))

	// 2. Fetch Real Price
	realPrice, err := p.fetcher.GetPrice(p.cfg.Symbol)
	if err != nil {
		zap.S().Errorw("Failed to fetch real price for reveal", "symbol", p.cfg.Symbol, "error", err)
		return
	}

	zap.S().Infow("Fetched real price", "symbol", p.cfg.Symbol, "price", realPrice)

	// 3. Pool Calculation
	// 3.1 Fetch and update system rollover
	// Use special userId "SYSTEM_POOL" and date "ROLLOVER" to store/retrieve rollover
	rolloverKey := SystemPoolUserId
	rolloverDate := RolloverDate
	sysRecord, err := dao.GetPricePredictDao().FindLatestByUser(rolloverKey)
	var rolloverPool int64 = 0
	if err == nil && sysRecord != nil && sysRecord.PredictDate == rolloverDate {
		rolloverPool = sysRecord.BetAmount
	} else {
		// Use RolloverDate as predictDate
		sysRecord = data.NewPricePredictData(rolloverKey, rolloverDate, 0)
		sysRecord.Status = data.PricePredictStatusRevealed
	}

	// 3.2 Categorize users and calculate current pool
	var winners []*data.PricePredictData
	var totalWinnerWeight float64 = 0
	var poolFromLosers float64 = 0.0
	var totalBetCurrent int64 = 0

	consolationRate := p.cfg.ConsolationRate
	if consolationRate <= 0 {
		consolationRate = 0.1 // Default 0.1
	}
	maxDiffLimit := p.cfg.MaxDiffLimit
	if maxDiffLimit <= 0 {
		maxDiffLimit = 5.0 // Default 5%
	}

	for _, user := range users {
		totalBetCurrent += user.BetAmount
		diffPercent := math.Abs(user.PredictPrice-realPrice) / realPrice * 100

		if diffPercent > maxDiffLimit {
			// Loser: Consolation Prize
			refund := float64(user.BetAmount) * consolationRate
			user.RewardPoints = int64(refund)
			user.Status = data.PricePredictStatusRevealed
			user.RealPrice = realPrice
			user.UpdateTime = time.Now().UTC().Unix()

			// Remaining goes to pool
			poolFromLosers += float64(user.BetAmount) * (1 - consolationRate)
		} else {
			// Winner: Qualify for pool
			score := 1.0 - (diffPercent / maxDiffLimit)
			if score < 0 {
				score = 0
			} // Should not happen given logic

			// Weight = Bet * (Score ^ 2)
			weight := float64(user.BetAmount) * math.Pow(score, 2)
			totalWinnerWeight += weight
			winners = append(winners, user)

			// Winner's bet effectively goes into the pool to be redistributed
			// We track it as part of the total distributable pool later
		}
	}

	// 3.3 Distribute
	distributablePool := float64(rolloverPool) + poolFromLosers
	// Add winners' bets to the pool (they share the whole pot)
	for _, w := range winners {
		distributablePool += float64(w.BetAmount)
	}

	if len(winners) == 0 {
		// No winners: Rollover everything
		// Losers already got their consolation recorded in user.RewardPoints (need to save below)
		rolloverPool = int64(distributablePool) // New rollover
		zap.S().Infow("No winners, pool rolls over", "amount", rolloverPool)
	} else {
		// Distribute to winners
		for _, w := range winners {
			diffPercent := math.Abs(w.PredictPrice-realPrice) / realPrice * 100
			score := 1.0 - (diffPercent / maxDiffLimit)
			weight := float64(w.BetAmount) * math.Pow(score, 2)

			ratio := weight / totalWinnerWeight
			reward := ratio * distributablePool
			w.RewardPoints = int64(reward)
			w.Status = data.PricePredictStatusRevealed
			w.RealPrice = realPrice
			w.UpdateTime = time.Now().UTC().Unix()
		}
		rolloverPool = 0 // Reset rollover
	}

	// 4. Save Updates
	// Save Users
	for _, user := range users {
		if err := dao.GetPricePredictDao().SaveOrUpdate(user); err != nil {
			zap.S().Errorw("Failed to save revealed user", "userId", user.Id, "error", err)
		}
	}
	// Save System Rollover
	sysRecord.BetAmount = rolloverPool
	sysRecord.RealPrice = realPrice
	sysRecord.UpdateTime = time.Now().UTC().Unix()
	if err := dao.GetPricePredictDao().SaveOrUpdate(sysRecord); err != nil {
		zap.S().Errorw("Failed to save system rollover", "error", err)
	}

	p.lastRevealDate = date

	// Reset Cache
	p.cacheMutex.Lock()
	p.cacheTodayPool = 0
	p.cacheUserCount = 0
	p.cacheMutex.Unlock()

	zap.S().Infow("Reveal completed", "winners", len(winners), "pool", distributablePool, "next_rollover", rolloverPool)
}

// CheckBetTime checks if the current time is within the betting window
func (p *PricePredictCenter) CheckBetTime() bool {
	if p.cfg == nil || !p.cfg.Open {
		return false
	}
	// Use UTC
	now := time.Now().UTC()
	var currentUnit int
	if config.IsDev() {
		currentUnit = now.Minute()
	} else {
		currentUnit = now.Hour()
	}
	return currentUnit >= int(p.cfg.BetStartTime) && currentUnit < int(p.cfg.BetEndTime)
}

func (p *PricePredictCenter) IsOpen() bool {
	return p.cfg != nil && p.cfg.Open
}

func (p *PricePredictCenter) GetTokenPrice(symbol string) (float64, error) {
	if p.fetcher == nil {
		return 0, fmt.Errorf("fetcher not initialized")
	}
	return p.fetcher.GetPrice(symbol)
}

type PredictPoolInfo struct {
	RolloverPool int64  `json:"rollover_pool"`
	TodayPool    int64  `json:"today_pool"`
	TotalPool    int64  `json:"total_pool"`
	UserCount    int64  `json:"user_count"`
	Symbol       string `json:"symbol"`
}

type PredictConfigInfo struct {
	Open            bool             `json:"open"`
	BetStartTime    int64            `json:"bet_start_time"`
	BetEndTime      int64            `json:"bet_end_time"`
	RevealTime      int64            `json:"reveal_time"`
	Symbol          string           `json:"symbol"`
	MaxDiffLimit    float64          `json:"max_diff_limit"`
	ConsolationRate float64          `json:"consolation_rate"`
	BetAmounts      []int64          `json:"bet_amounts"`
	PoolInfo        *PredictPoolInfo `json:"pool_info"`
}

func (p *PricePredictCenter) GetConfigInfo() (*PredictConfigInfo, error) {
	if p.cfg == nil {
		return nil, fmt.Errorf("price predict not configured")
	}
	// Fetch current pool info
	rolloverKey := SystemPoolUserId
	sysRecord, err := dao.GetPricePredictDao().FindLatestByUser(rolloverKey)
	var rolloverPool int64 = 0
	if err == nil && sysRecord != nil {
		rolloverPool = sysRecord.BetAmount
	}

	// Fetch today's betting stats (From Cache)
	p.cacheMutex.RLock()
	todayPool := p.cacheTodayPool
	userCount := p.cacheUserCount
	p.cacheMutex.RUnlock()

	// Calculate Timestamps (UTC)
	now := time.Now().UTC()
	var betStartTs, betEndTs, revealTs int64

	if config.IsDev() {
		// Dev: Base is current Hour start
		baseTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		betStartTs = baseTime.Add(time.Duration(p.cfg.BetStartTime) * time.Minute).Unix()
		betEndTs = baseTime.Add(time.Duration(p.cfg.BetEndTime) * time.Minute).Unix()
		revealTs = baseTime.Add(time.Duration(p.cfg.RevealTime) * time.Minute).Unix()
	} else {
		// Prod: Base is current Day start
		toDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		betStartTs = toDayStart.Add(time.Duration(p.cfg.BetStartTime) * time.Hour).Unix()
		betEndTs = toDayStart.Add(time.Duration(p.cfg.BetEndTime) * time.Hour).Unix()
		revealTs = toDayStart.Add(time.Duration(p.cfg.RevealTime) * time.Hour).Unix()
	}

	return &PredictConfigInfo{
		Open:            p.cfg.Open,
		BetStartTime:    betStartTs,
		BetEndTime:      betEndTs,
		RevealTime:      revealTs,
		Symbol:          p.cfg.Symbol,
		MaxDiffLimit:    p.cfg.MaxDiffLimit,
		ConsolationRate: p.cfg.ConsolationRate,
		BetAmounts: func() []int64 {
			if len(p.cfg.BetAmounts) > 0 {
				return p.cfg.BetAmounts
			}
			return []int64{100, 200, 500, 1000}
		}(),
		PoolInfo: &PredictPoolInfo{
			RolloverPool: rolloverPool,
			TodayPool:    todayPool,
			TotalPool:    rolloverPool + todayPool,
			UserCount:    userCount,
			Symbol:       p.cfg.Symbol,
		},
	}, nil
}

func (p *PricePredictCenter) CreateOrUpdateUserPredict(userId string, price float64, betAmount int64) (*data.PricePredictData, error) {
	// 1. Check latest record status
	record, err := dao.GetPricePredictDao().FindLatestByUser(userId)

	now := time.Now().UTC()
	date := p.getPredictDate(now)

	// Single Record Logic:
	// - If record exists:
	//   - If Status == Claimed: ALLOW overwrite (Start new round).
	//   - If Status == Unsubmitted: ALLOW overwrite/update.
	//   - If Status == Pending:
	//       - If Date == CurrentDate: ALLOW update (Price only).
	//       - If Date != CurrentDate: ERROR (Previous round pending, wait for reveal).
	//   - If Status == Revealed: ERROR (Please claim reward first).

	var targetRecord *data.PricePredictData
	isNewRound := false

	if err == nil && record != nil {
		switch record.Status {
		case data.PricePredictStatusClaimed:
			// Reset for new round
			targetRecord = record
			targetRecord.PredictDate = date
			targetRecord.PredictPrice = price
			targetRecord.Status = data.PricePredictStatusPending
			targetRecord.BetAmount = betAmount
			targetRecord.RewardPoints = 0
			targetRecord.CreateTime = now.Unix()
			targetRecord.UpdateTime = now.Unix()
			isNewRound = true
		case data.PricePredictStatusRevealed:
			return nil, fmt.Errorf("please claim your previous reward first")
		case data.PricePredictStatusPending:
			if record.PredictDate == date {
				// Update existing pending bet
				targetRecord = record
				isNewRound = false // Not new, just update
			} else {
				return nil, fmt.Errorf("previous round pending, please wait for result")
			}
		case data.PricePredictStatusUnsubmitted:
			// Should act as new
			targetRecord = record
			targetRecord.PredictDate = date // Ensure date is current
			targetRecord.PredictPrice = price
			targetRecord.Status = data.PricePredictStatusPending
			targetRecord.BetAmount = betAmount
			targetRecord.CreateTime = now.Unix()
			targetRecord.UpdateTime = now.Unix()
			isNewRound = true
		}
	} else {
		// No record found, create new
		targetRecord = data.NewPricePredictData(userId, date, price)
		targetRecord.Status = data.PricePredictStatusPending
		targetRecord.BetAmount = betAmount
		isNewRound = true
	}

	// 2. Check bet time (Must be done before Cost)
	if !p.CheckBetTime() {
		return nil, fmt.Errorf("not in betting time")
	}

	// 3. Check bet amount validity (if new round)
	if isNewRound {
		allowedAmounts := p.cfg.BetAmounts
		if len(allowedAmounts) == 0 {
			allowedAmounts = []int64{100, 200, 500, 1000}
		}
		if !slices.Contains(allowedAmounts, betAmount) {
			return nil, fmt.Errorf("invalid bet amount, allowed: %v", allowedAmounts)
		}

		// 4. Deduct Balance
		// Use "NolanDevPoint" as requested
		c, existed := coincenter.Get().GetCoinByName("NolanDevPoint")
		if !existed {
			return nil, fmt.Errorf("bet coin not found")
		}

		// DecUserCoinAmount is atomic and thread-safe
		_, err := coincenter.Get().DecUserCoinAmount(userId, c.Id, int32(betAmount), "price_predict_bet", targetRecord.Id)
		if err != nil {
			return nil, fmt.Errorf("insufficient balance")
		}

		// Update Cache
		p.cacheMutex.Lock()
		p.cacheTodayPool += betAmount
		p.cacheUserCount++
		p.cacheMutex.Unlock()
	} else {
		// Update existing: Check if attempting to change amount
		if targetRecord.BetAmount != betAmount {
			// For simplicity, disallow changing bet amount in same round
			return nil, fmt.Errorf("cannot change bet amount for existing bet")
		}
		targetRecord.PredictPrice = price
		targetRecord.UpdateTime = now.Unix()
	}

	// 5. Save
	if err := dao.GetPricePredictDao().SaveOrUpdate(targetRecord); err != nil {
		zap.S().Errorw("Failed to save prediction", "userId", userId, "error", err)
		return nil, err
	}

	return targetRecord, nil
}

func (p *PricePredictCenter) GetUserPredict(userId string) (*data.PricePredictData, error) {
	record, err := dao.GetPricePredictDao().FindLatestByUser(userId)
	// If not found (err != nil) or status is Claimed, return a new Unsubmitted data
	if err != nil || (record != nil && record.Status == data.PricePredictStatusClaimed) {
		date := p.getPredictDate(time.Now().UTC())
		return data.NewPricePredictData(userId, date, 0), nil
	}
	return record, nil
}

func (p *PricePredictCenter) ClaimReward(userId string) (*data.PricePredictData, error) {
	// 1. Get latest record
	record, err := dao.GetPricePredictDao().FindLatestByUser(userId)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("no prediction record found")
	}

	// 2. Check status
	if record.Status == data.PricePredictStatusClaimed {
		return nil, fmt.Errorf("reward already claimed")
	}
	if record.Status != data.PricePredictStatusRevealed {
		return nil, fmt.Errorf("reward not ready")
	}

	// 3. Update status
	record.Status = data.PricePredictStatusClaimed
	record.UpdateTime = time.Now().Unix()

	// 4. Save
	if err := dao.GetPricePredictDao().SaveOrUpdate(record); err != nil {
		return nil, err
	}

	// 5. Add points to user wallet
	if record.RewardPoints > 0 {
		// Reward NolanDevPoint (as manually requested)
		if c, existed := coincenter.Get().GetCoinByName("NolanDevPoint"); existed {
			coincenter.Get().AddUserCoinAmount(userId, c.Id, int32(record.RewardPoints), "price_predict_reward", record.Id)
		}
	}
	zap.S().Infow("User claimed reward", "userId", userId, "points", record.RewardPoints)

	return record, nil
}

func (p *PricePredictCenter) getBetCoin() (*data.CoinData, error) {
	c, existed := coincenter.Get().GetCoinByName("NolanDevPoint")
	if !existed {
		return nil, fmt.Errorf("bet coin not found")
	}
	// CoinCenter returns *data.CoinProto usually but let's check coincenter.go to be sure.
	// Actually CoinCenter.GetCoinByName returns (*data.CoinProto, bool).
	// If the compiler says undefined: data.CoinProto, it means we need to see where CoinProto is defined.
	// It's likely in strictly typed generated proto files or similar.
	// Let's check coincenter.go first.
	return c, nil
}

// getCurrentHashPrice gets the current price from fetcher for bots
func (p *PricePredictCenter) getCurrentHashPrice() float64 {
	price, _ := p.fetcher.GetPrice(p.cfg.Symbol)
	return price
}
