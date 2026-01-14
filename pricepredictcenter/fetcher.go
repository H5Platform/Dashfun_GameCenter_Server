package pricepredictcenter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type PriceFetcher struct {
	client     *http.Client
	cachePrice float64
	cacheTime  int64
	mu         sync.Mutex
}

func NewPriceFetcher() *PriceFetcher {
	return &PriceFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetPrice fetches price from multiple sources and returns the median/consensus price
func (f *PriceFetcher) GetPrice(symbol string) (float64, error) {
	// Check Cache
	f.mu.Lock()
	now := time.Now().Unix()
	if now-f.cacheTime < 5 && f.cachePrice > 0 {
		price := f.cachePrice
		f.mu.Unlock()
		return price, nil
	}
	f.mu.Unlock()

	symbol = strings.ToUpper(symbol)
	var prices []float64
	var mu sync.Mutex
	var wg sync.WaitGroup

	sources := []func(string) (float64, error){
		f.fetchBinance,
		f.fetchCoinGecko,
		f.fetchOKX,
	}

	for _, source := range sources {
		wg.Add(1)
		go func(s func(string) (float64, error)) {
			defer wg.Done()
			price, err := s(symbol)
			if err == nil && price > 0 {
				mu.Lock()
				prices = append(prices, price)
				mu.Unlock()
			} else {
				zap.S().Warnw("Failed to fetch price", "symbol", symbol, "error", err)
			}
		}(source)
	}

	wg.Wait()

	if len(prices) == 0 {
		return 0, fmt.Errorf("failed to fetch price from all sources for %s", symbol)
	}

	// Calculate median or consensus
	// If we have 3 prices, sort and take middle.
	// If 2, average.
	// If 1, return it.
	sort.Float64s(prices)

	zap.S().Infow("Fetched prices", "symbol", symbol, "prices", prices)

	var result float64
	n := len(prices)
	if n%2 == 1 {
		result = prices[n/2]
	} else {
		result = (prices[n/2-1] + prices[n/2]) / 2.0
	}

	// Update Cache
	f.mu.Lock()
	f.cachePrice = result
	f.cacheTime = now
	f.mu.Unlock()

	return result, nil
}

func (f *PriceFetcher) fetchBinance(symbol string) (float64, error) {
	// API: https://api.binance.com/api/v3/ticker/price?symbol=BNBUSDT
	// Note: Binance requires specific pair format. Assuming USDT pair for now.
	pair := symbol + "USDT"
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", pair)

	resp, err := f.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("binance status %d", resp.StatusCode)
	}

	var result struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	var price float64
	if _, err := fmt.Sscanf(result.Price, "%f", &price); err != nil {
		return 0, err
	}
	return price, nil
}

func (f *PriceFetcher) fetchCoinGecko(symbol string) (float64, error) {
	// API: https://api.coingecko.com/api/v3/simple/price?ids=binancecoin&vs_currencies=usd
	// CoinGecko uses IDs, not symbols (e.g. 'binancecoin' for BNB).
	// This is tricky. We might need a mapper or config.
	// simpler hack: try to map common ones or search?
	// User requirement: "look at config symbol".
	// If config says "BNB", CoinGecko expects "binancecoin".
	// For now, let's include a small map or fallback.

	id := strings.ToLower(symbol)
	if id == "bnb" {
		id = "binancecoin"
	} else if id == "btc" {
		id = "bitcoin"
	} else if id == "eth" {
		id = "ethereum"
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", id)

	resp, err := f.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coingecko status %d", resp.StatusCode)
	}

	// Response: {"binancecoin":{"usd":623.45}}
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if val, ok := result[id]; ok {
		return val["usd"], nil
	}
	return 0, fmt.Errorf("id not found in coingecko response")
}

func (f *PriceFetcher) fetchOKX(symbol string) (float64, error) {
	// API: https://www.okx.com/api/v5/market/ticker?instId=BNB-USDT
	pair := fmt.Sprintf("%s-USDT", strings.ToUpper(symbol))
	url := fmt.Sprintf("https://www.okx.com/api/v5/market/ticker?instId=%s", pair)

	resp, err := f.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("okx status %d", resp.StatusCode)
	}

	// Response: {"code":"0", "data": [{"last": "..."}]}
	var result struct {
		Code string `json:"code"`
		Data []struct {
			Last string `json:"last"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Code != "0" || len(result.Data) == 0 {
		return 0, fmt.Errorf("okx api error or no data")
	}

	var price float64
	if _, err := fmt.Sscanf(result.Data[0].Last, "%f", &price); err != nil {
		return 0, err
	}
	return price, nil
}
