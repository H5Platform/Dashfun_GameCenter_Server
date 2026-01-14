package pricepredictcenter

import (
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/dao"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type PricePredictBot struct {
	cfg                *config.PricePredictConfig
	worker             *snowflake.Worker
	botIds             []string
	center             *PricePredictCenter
	stopChan           chan struct{}
	lastReveal         int64 // Track last reveal time to detect new rounds
	currentCycleDate   string
	currentCycleTarget int
}

func NewPricePredictBot(center *PricePredictCenter) *PricePredictBot {
	w, err := snowflake.GetWorker(data.WorkerPredictBotId)
	if err != nil {
		zap.S().Panicw("Failed to init bot worker", "error", err)
	}

	return &PricePredictBot{
		cfg:      center.cfg,
		worker:   w,
		center:   center,
		stopChan: make(chan struct{}),
		botIds:   make([]string, 0),
	}
}

func (b *PricePredictBot) Start() {
	go b.run()
}

func (b *PricePredictBot) Stop() {
	close(b.stopChan)
}

func (b *PricePredictBot) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial load of existing bots
	b.loadBots()

	for {
		select {
		case <-b.stopChan:
			return
		case <-ticker.C:
			b.checkAndBet()
		}
	}
}

func (b *PricePredictBot) loadBots() {
	// Load existing bots from DB to reuse IDs
	// Limit to MaxBotCount * 2 to have some buffer, though we strictly respect MaxBotCount for active betting
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	limit := int64(b.cfg.MaxBotCount)
	if limit <= 0 {
		limit = 500 // Default fallback
	}

	bots, err := dao.GetPricePredictDao().FindBots(ctx, limit)
	if err != nil {
		zap.S().Errorw("Failed to load bots", "error", err)
		return
	}

	b.botIds = make([]string, 0, len(bots))
	seen := make(map[string]bool)
	for _, bot := range bots {
		// Only track unique UserIDs
		if !seen[bot.Id] {
			b.botIds = append(b.botIds, bot.Id)
			seen[bot.Id] = true
		}
	}
	zap.S().Infow("Loaded existing bots", "count", len(b.botIds))
}

func (b *PricePredictBot) checkAndBet() {
	// 1. Check if betting time
	if !b.center.CheckBetTime() {
		return
	}

	now := time.Now().UTC()
	todayDate := b.center.getPredictDate(now)

	// Check for new cycle and randomize target
	if todayDate != b.currentCycleDate {
		baseCount := b.cfg.MaxBotCount
		// +/- 20%
		variationPercent := (rand.Float64() * 0.4) - 0.2 // -0.2 to +0.2
		variation := int(float64(baseCount) * variationPercent)
		newTarget := max(baseCount+variation, 0)

		b.currentCycleTarget = newTarget
		b.currentCycleDate = todayDate
		zap.S().Infow("Bot cycle reset", "date", todayDate, "target", newTarget, "base", baseCount)
	}

	// 2. Linear Distribution Strategy
	// Calculate expected number of bots that should have bet by now

	betStart := b.cfg.BetStartTime
	betEnd := b.cfg.BetEndTime

	var totalWindowMin, elapsedMin int

	if config.IsDev() {
		// Dev Mode: Config times are Minutes (0-59) within the hour
		// Window: [StartMin, EndMin]
		totalWindowMin = betEnd - betStart
		currentMinute := now.Minute()
		elapsedMin = currentMinute - betStart
	} else {
		// Prod Mode: Config times are Hours (0-23) within the day
		// Window: [StartHour*60, EndHour*60]
		// Assuming Start < End for simplicity as per current config
		startTotalMin := betStart * 60
		endTotalMin := betEnd * 60
		totalWindowMin = endTotalMin - startTotalMin

		currentHour := now.Hour()
		currentMinute := now.Minute()
		currentTotalMin := currentHour*60 + currentMinute
		elapsedMin = currentTotalMin - startTotalMin
	}

	if totalWindowMin <= 0 {
		return // Invalid config or instant close
	}

	if elapsedMin < 0 {
		return // Should be caught by CheckBetTime but double check
	}

	// Progress ratio (0.0 to 1.0)
	progress := float64(elapsedMin) / float64(totalWindowMin)
	if progress > 1.0 {
		progress = 1.0
	}

	targetCount := int(float64(b.currentCycleTarget) * progress)

	// 3. Check how many bots have ALREADY bet today
	// We assume fetcher or DAO availability.
	// Ideally we need a distinct count of bots for today.
	// Since DAO doesn't have "CountBotsByDate", we might need to rely on local state or add DAO method.
	// For simplicity and robustness, let's query the pending predictions for today and count IsBot.
	// This might be heavy if many predictions.
	// Optimization: Cache this in 'b.center' or 'b' if possible?
	// Let's implement a 'FindPendingPredictionsByDate' filtering in memory for now (since we fetch all Pending anyway for reveal)
	// But Reveal fetches at the END. We need it during the day.

	// Let's rely on checking our known botIds vs database status one by one? Too slow.
	// Alternative: Add 'CountBotsByDate' to DAO.
	// For now, let's just bet if 'random' condition met? No, user asked for distribution.

	// Let's Assume we can query DAO or reuse 'FindPendingPredictionsByDate'.
	pendings, err := dao.GetPricePredictDao().FindPendingPredictionsByDate(todayDate)
	if err != nil {
		zap.S().Errorw("Bot failed to check pending count", "error", err)
		return
	}

	currentBotCount := 0
	existingBetters := make(map[string]bool)
	for _, p := range pendings {
		if p.IsBot {
			currentBotCount++
			existingBetters[p.Id] = true // Record who has bet
		}
	}

	needed := targetCount - currentBotCount
	if needed <= 0 {
		return // Ahead of schedule
	}

	// 4. Fire 'needed' bots
	zap.S().Infow("Bot triggering bets", "needed", needed, "progress", progress)

	for i := 0; i < needed; i++ {
		// Find a bot that hasn't bet yet
		var botId string

		// Try to find available existing bot
		// Shuffle usage? Or linear scan with offset?
		// Random pick to avoid pattern
		availableIndices := rand.Perm(len(b.botIds))
		found := false
		for _, idx := range availableIndices {
			id := b.botIds[idx]
			if !existingBetters[id] {
				botId = id
				found = true
				break
			}
		}

		// If no existing bot available and we haven't reached MaxLimit totally (globally), create new
		if !found {
			if len(b.botIds) < b.cfg.MaxBotCount {
				// Generate new
				newId := "pp_bot_" + b.worker.NextStrId()
				b.botIds = append(b.botIds, newId)
				botId = newId
			} else {
				// We ran out of bots? Should not happen if logic correct (MaxBotCount consistency)
				// Just pick any to overwrite? No, one bet per round.
				// Stop trying if no bots available
				break
			}
		}

		b.execBet(botId, todayDate)
		existingBetters[botId] = true // Mark as used
	}
}

func (b *PricePredictBot) execBet(botId, date string) {
	// 5. Generate Bet Data

	// Price
	// Get current price from fetcher (need access? It's private in center?)
	// Center calls fetcher.fetch(). We can expose a GetCurrentPrice() on center or fetcher.
	// Assuming api.GetMarketPrice is available or we use fetcher.
	// Let's reuse api.GetMarketPrice from 'api_markets.go' or similar?
	// Wait, we implemented 'fetcher' in pricepredictcenter.
	// We can add a helper in Center.

	// Temporary direct fetch or helper
	// Assuming center has a way to get price.
	// If not, we'll implement a simple one here or make fetcher public.
	// Let's use `api.ApiGetSymbolPrice` if available? No, layers.
	// Let's assume we can get it from Center if we add a getter.
	// modifying center to expose getCurrentPrice.

	// Check b.center.fetcher
	currentPrice := b.center.getCurrentHashPrice() // We need to add this method or similar
	if currentPrice <= 0 {
		// Fallback or skip
		return
	}

	// Randomize +/- 10%
	variation := (rand.Float64() * 0.2) - 0.1 // -0.1 to +0.1
	predictPrice := currentPrice * (1 + variation)

	// Amount
	amounts := b.cfg.BetAmounts
	if len(amounts) == 0 {
		amounts = []int64{100, 200, 500, 1000}
	}
	amount := amounts[rand.Intn(len(amounts))]

	// Create Data
	// NEW: Force Status to Revealed to avoid claiming?
	// User said: "机器人不领取奖励... 直接修改状态" -> "Bots don't claim... modify status directly"
	// But it also says "randomly distributed IN betting cycle".
	// If we set status to Revealed NOW, it won't be counted in the "Pending" list for the Reveal Logic?
	// The Reveal logic queries "FindPendingPredictionsByDate".
	// IF we set Status=Revealed immediately, they will be IGNORED by Reveal Result calculation.
	// BUT User said: "机器人数据需要能区分出来... 方便每一轮直接读取机器人数据进行模拟... 机器人不领取奖励"
	// If they are not in Reveal, they don't affect the pool (Weighted Average).
	// "负责在下注时间内随机下注... 需要在每个周期内完成这个上限人数"
	// "机器人不领取奖励" usually means they participate in the GAME (affecting pool/odds?) but don't physically act to claim.
	// IF they need to affect the game (pool size, winner calculation), they MUST be PENDING.
	// Once Revealed, they stay Revealed (don't move to Claimed).
	// SO: Create as PENDING.
	// "直接修改状态" might refer to skipping the 'Claim' user action state transition.

	ppData := &data.PricePredictData{
		Id:           botId,
		PredictDate:  date,
		PredictPrice: predictPrice,
		RealPrice:    0,
		IsBot:        true,
		Status:       data.PricePredictStatusPending,
		BetAmount:    amount,
		RewardPoints: 0,
		CreateTime:   time.Now().UTC().Unix(),
		UpdateTime:   time.Now().UTC().Unix(),
	}

	// Save directly
	if err := dao.GetPricePredictDao().SaveOrUpdate(ppData); err != nil {
		zap.S().Errorw("Bot failed to save bet", "id", botId, "error", err)
		return
	}

	// Update Cache manually
	b.center.cacheMutex.Lock()
	b.center.cacheTodayPool += amount
	b.center.cacheUserCount++
	b.center.cacheMutex.Unlock()
}
