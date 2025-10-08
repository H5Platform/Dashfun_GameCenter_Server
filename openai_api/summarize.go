package openai_api

import (
	"dashfun_gamecenter/coingecko"
	"go.uber.org/zap"
	"sync"
	"time"
)

var once sync.Once
var inst *MarketSummarize

type summarize struct {
	content    string
	updateTime time.Time
}
type MarketSummarize struct {
	summarizes map[string]*summarize
	mu         sync.RWMutex
}

func GetMarketSummarize() *MarketSummarize {
	once.Do(func() {
		inst = &MarketSummarize{}
		inst.init()
	})
	return inst
}

func (s *MarketSummarize) init() {
	s.summarizes = make(map[string]*summarize)
	s.StartAutoUpdate(30 * time.Minute)
}

func (s *MarketSummarize) updateSummarize(tokenId string) {
	prompt, err := coingecko.GetTokenPricePrompt(tokenId, "usd")
	if err != nil {
		zap.S().Errorw("get token price prompt failed", "err", err)
		return
	}
	summary, err := SummarizeWithOpenAI(prompt)
	if err != nil {
		zap.S().Errorw("summarize with openai failed", "err", err)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summarizes[tokenId] = &summarize{
		content:    summary,
		updateTime: time.Now(),
	}
}

func (s *MarketSummarize) updateSummarizeTick() {
	for _, tokenId := range coingecko.Get().KnownTokens() {
		s.updateSummarize(tokenId)
	}
}

func (s *MarketSummarize) StartAutoUpdate(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			s.updateSummarizeTick()
			<-ticker.C
		}
	}()
}

func (s *MarketSummarize) GetSummarize(tokenId string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sum, ok := s.summarizes[tokenId]; ok {
		return sum.content
	}
	return ""
}
