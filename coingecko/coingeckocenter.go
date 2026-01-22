package coingecko

import (
	"dashfun_gamecenter/config"
	"sync"
	"time"

	"go.uber.org/zap"
)

var once sync.Once
var inst *CoinGeckoCenter

type CoinGeckoCenter struct {
	knownTokens []string //已知的token id列表
	marketInfo  map[string]*TokenMarketInfo
	lastUpdated time.Time
	mu          sync.RWMutex
}

func Get() *CoinGeckoCenter {
	once.Do(func() {
		inst = &CoinGeckoCenter{}
		inst.init()
	})
	return inst
}

func (c *CoinGeckoCenter) init() {
	cfg := config.GetConfig().CoinGeckoConfig
	c.marketInfo = make(map[string]*TokenMarketInfo)
	c.knownTokens = cfg.DefaultTokenIds
	if len(c.knownTokens) == 0 {
		c.knownTokens = []string{"ethereum", "bitcoin"}
	}
	c.lastUpdated = time.Time{}

	interval := cfg.UpdateInterval
	if interval == 0 {
		interval = 1
	}

	c.StartAutoUpdate(time.Duration(interval) * time.Minute)
}

// StartAutoUpdate 定时更新市场信息
func (c *CoinGeckoCenter) StartAutoUpdate(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			c.updateMarkets()
			<-ticker.C
		}
	}()
}

// updateMarkets 更新市场信息
func (c *CoinGeckoCenter) updateMarkets() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.knownTokens) == 0 {
		return
	}
	results, err := fetchMarketsMany(c.knownTokens, "usd")
	if err != nil {
		zap.S().Errorw("coingeckoCenter.updateMarkets Error", "err", err)
		return
	}
	for id, info := range results {
		c.marketInfo[id] = info
	}
	c.lastUpdated = time.Now()
}

// GetMarketInfo 获取指定id的TokenMarketInfo
func (c *CoinGeckoCenter) GetMarketInfo(id string) (*TokenMarketInfo, error) {
	c.mu.RLock()
	info, ok := c.marketInfo[id]
	c.mu.RUnlock()
	if ok {
		return info, nil
	}

	// 未命中，写锁
	c.mu.Lock()
	defer c.mu.Unlock()
	// 双重检查，防止并发重复写
	info, ok = c.marketInfo[id]
	if ok {
		return info, nil
	}
	result, err := fetchMarkets(id, "usd")
	if err != nil {
		zap.S().Errorw("coingeckoCenter.GetMarketInfo fetchMarkets Error", "id", id, "err", err)
		return nil, err
	}
	c.marketInfo[id] = result
	return result, nil
}

func (c *CoinGeckoCenter) KnownTokens() []string {
	return c.knownTokens
}
