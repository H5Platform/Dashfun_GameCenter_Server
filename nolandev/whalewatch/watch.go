package whalewatch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ======== 常量 / 配置 ========

// ERC20 Transfer topic0 = keccak256("Transfer(address,address,uint256)")
var topicTransfer = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

type Chain struct {
	Name           string                      // "ethereum" | "bsc" | "polygon"
	WSURL          string                      // wss://...
	NativeSymbol   string                      // ETH | BNB | MATIC
	NativeCGID     string                      // coingecko id: ethereum | binancecoin | matic-network
	NativeDecimals int                         // 18
	USDT           common.Address              // USDT 合约地址（各链不同）
	USDTDecimals   int                         // 6
	Ignore         map[common.Address]struct{} // 可选：忽略地址集（CEX等）
	MinNativeUSD   float64                     // 原生币报警阈值（USD）
	MinUSDTUSD     float64                     // USDT 报警阈值（USD）
}

// 三链默认配置（地址已填好）
func DefaultChains(wsETH, wsBSC, wsPOLY string, minNativeUSD, minUSDTUSD float64,
	ignoreETH, ignoreBSC, ignorePOLY []string) []Chain {

	return []Chain{
		{
			Name:           "ethereum",
			WSURL:          wsETH,
			NativeSymbol:   "ETH",
			NativeCGID:     "ethereum",
			NativeDecimals: 18,
			USDT:           common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
			USDTDecimals:   6,
			Ignore:         toAddrSet(ignoreETH),
			MinNativeUSD:   minNativeUSD,
			MinUSDTUSD:     minUSDTUSD,
		},
		{
			Name:           "bsc",
			WSURL:          wsBSC,
			NativeSymbol:   "BNB",
			NativeCGID:     "binancecoin",
			NativeDecimals: 18,
			USDT:           common.HexToAddress("0x55d398326f99059fF775485246999027B3197955"),
			USDTDecimals:   6,
			Ignore:         toAddrSet(ignoreBSC),
			MinNativeUSD:   minNativeUSD,
			MinUSDTUSD:     minUSDTUSD,
		},
		{
			Name:           "polygon",
			WSURL:          wsPOLY,
			NativeSymbol:   "MATIC",
			NativeCGID:     "matic-network",
			NativeDecimals: 18,
			USDT:           common.HexToAddress("0xc2132D05D31c914a87C6611C10748AEb04B58e8F"),
			USDTDecimals:   6,
			Ignore:         toAddrSet(ignorePOLY),
			MinNativeUSD:   minNativeUSD,
			MinUSDTUSD:     minUSDTUSD,
		},
	}
}

func toAddrSet(list []string) map[common.Address]struct{} {
	if len(list) == 0 {
		return nil
	}
	m := make(map[common.Address]struct{}, len(list))
	for _, s := range list {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m[common.HexToAddress(s)] = struct{}{}
	}
	return m
}

// ======== 简易价格缓存（CoinGecko simple/price） ========
type priceCache struct {
	mu       sync.Mutex
	last     map[string]float64
	lastTime time.Time
	ttl      time.Duration
	client   *http.Client
}

func newPriceCache(ttl time.Duration) *priceCache {
	return &priceCache{
		last:   make(map[string]float64),
		ttl:    ttl,
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (p *priceCache) getUSD(ids ...string) (map[string]float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	needFetch := now.Sub(p.lastTime) > p.ttl
	if !needFetch {
		for _, id := range ids {
			if _, ok := p.last[id]; !ok {
				needFetch = true
				break
			}
		}
	}

	if needFetch {
		url := "https://api.coingecko.com/api/v3/simple/price?vs_currencies=usd&ids=" + strings.Join(ids, ",")
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "whale-watch/1.0")
		resp, err := p.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var m map[string]map[string]float64
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			return nil, err
		}
		for k, v := range m {
			p.last[k] = v["usd"]
		}
		p.lastTime = now
	}

	out := make(map[string]float64, len(ids))
	for _, id := range ids {
		out[id] = p.last[id]
	}
	return out, nil
}

// ======== 监听：USDT Transfer（事件） ========
func runUSDTWatcher(ctx context.Context, c Chain, client *ethclient.Client, prices *priceCache, logf func(format string, a ...any)) error {
	if c.USDT == (common.Address{}) || c.MinUSDTUSD <= 0 {
		return nil
	}
	q := ethereum.FilterQuery{
		Addresses: []common.Address{c.USDT},
		Topics:    [][]common.Hash{{topicTransfer}},
	}
	logs := make(chan types.Log, 2048)
	sub, err := client.SubscribeFilterLogs(ctx, q, logs)
	if err != nil {
		return fmt.Errorf("[%s][USDT] subscribe logs: %w", c.Name, err)
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				logf("[%s][USDT] subscribe error: %v", c.Name, err)
				return
			case ev := <-logs:
				if len(ev.Topics) < 3 || len(ev.Data) < 32 {
					continue
				}
				from := common.BytesToAddress(ev.Topics[1].Bytes())
				to := common.BytesToAddress(ev.Topics[2].Bytes())
				if inSet(from, c.Ignore) || inSet(to, c.Ignore) {
					continue
				}
				val := new(big.Int).SetBytes(ev.Data[len(ev.Data)-32:]) // uint256
				amt := toFloat(val, c.USDTDecimals)                     // USDT数量（十进制）
				pm, err := prices.getUSD("tether")
				if err != nil {
					logf("[%s][USDT] price err: %v", c.Name, err)
					continue
				}
				usd := amt * pm["tether"]
				if usd >= c.MinUSDTUSD {
					logf("[WHALE][%s][USDT] $%.0f | %.0f USDT | %s -> %s | tx=%s | block=%d",
						strings.ToUpper(c.Name), usd, math.Round(amt),
						from.Hex(), to.Hex(), ev.TxHash.Hex(), ev.BlockNumber)
				}
			}
		}
	}()

	return nil
}

// ======== 监听：原生币大额转账（新块扫描） ========
func runNativeWatcher(ctx context.Context, c Chain, client *ethclient.Client, prices *priceCache, logf func(format string, a ...any)) error {
	if c.MinNativeUSD <= 0 {
		return nil
	}
	headers := make(chan *types.Header, 256)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return fmt.Errorf("[%s][NATIVE] newHead: %w", c.Name, err)
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				logf("[%s][NATIVE] newHead error: %v", c.Name, err)
				return
			case h := <-headers:
				blk, err := client.BlockByHash(ctx, h.Hash())
				if err != nil {
					logf("[%s][NATIVE] get block: %v", c.Name, err)
					continue
				}
				pm, err := prices.getUSD(c.NativeCGID)
				if err != nil {
					logf("[%s][NATIVE] price err: %v", c.Name, err)
					continue
				}
				priceUSD := pm[c.NativeCGID]

				for _, tx := range blk.Transactions() {
					to := ""
					if tx.To() != nil {
						to = tx.To().Hex()
						if inSet(*tx.To(), c.Ignore) {
							continue
						}
					}
					// from
					from := "unknown"
					signer := types.LatestSignerForChainID(tx.ChainId())
					fromAddr, err := types.Sender(signer, tx)
					if err == nil {
						if inSet(fromAddr, c.Ignore) {
							continue
						}
						from = fromAddr.Hex()
					} else {
						from = "unknown"
					}

					val := tx.Value()
					if val == nil || val.Sign() == 0 {
						continue
					}
					amt := toFloat(val, c.NativeDecimals)
					usd := amt * priceUSD
					if usd >= c.MinNativeUSD {
						logf("[WHALE][%s][%s] $%.0f | %.6f %s | %s -> %s | tx=%s | block=%d",
							strings.ToUpper(c.Name), c.NativeSymbol, usd, amt, c.NativeSymbol,
							from, to, tx.Hash().Hex(), blk.NumberU64())
					}
				}
			}
		}
	}()

	return nil
}

func inSet(a common.Address, set map[common.Address]struct{}) bool {
	if set == nil {
		return false
	}
	_, ok := set[a]
	return ok
}

func toFloat(val *big.Int, decimals int) float64 {
	if val == nil {
		return 0
	}
	scale := new(big.Float).SetFloat64(math.Pow10(decimals))
	f := new(big.Float).Quo(new(big.Float).SetInt(val), scale)
	out, _ := f.Float64()
	return out
}

// ======== 入口：并行跑多链 ========
type Logger func(format string, a ...any)

// Run 会为每条链启动原生币 & USDT 两个监听协程。
// 你可以传自定义 log 函数（例如写到文件/DB），不传则用标准 log.Printf。
func Run(ctx context.Context, chains []Chain, loggers ...Logger) error {
	var logf Logger = func(format string, a ...any) { log.Printf(format, a...) }
	if len(loggers) > 0 && loggers[0] != nil {
		logf = loggers[0]
	}

	// 归并出所有需要的 CoinGecko id 一次性缓存
	idSet := map[string]struct{}{}
	for _, c := range chains {
		if c.WSURL == "" {
			continue
		}
		idSet[c.NativeCGID] = struct{}{}
		idSet["tether"] = struct{}{}
	}
	if len(idSet) == 0 {
		return fmt.Errorf("no active chains (missing WSURL)")
	}
	var ids []string
	for id := range idSet {
		ids = append(ids, id)
	}

	prices := newPriceCache(30 * time.Second)
	// 预热
	if _, err := prices.getUSD(ids...); err != nil {
		logf("price warmup err: %v", err)
	}

	// 每条链连接并启动 watcher
	for _, c := range chains {
		if c.WSURL == "" {
			continue
		}
		c := c
		go func() {
			client, err := ethclient.DialContext(ctx, c.WSURL)
			if err != nil {
				logf("[%s] dial err: %v", c.Name, err)
				return
			}
			defer client.Close()

			if err := runUSDTWatcher(ctx, c, client, prices, logf); err != nil {
				logf("%v", err)
			}
			if err := runNativeWatcher(ctx, c, client, prices, logf); err != nil {
				logf("%v", err)
			}
			logf("[%s] whale watchers started (native:%s & USDT)", strings.ToUpper(c.Name), c.NativeSymbol)
		}()
	}
	return nil
}
