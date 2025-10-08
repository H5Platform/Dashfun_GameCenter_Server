package coingecko

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// ===== CoinGecko: /coins/markets 响应结构 =====
type TokenMarketInfo struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	High24h           float64 `json:"high_24h"`
	Low24h            float64 `json:"low_24h"`
	MarketCap         float64 `json:"market_cap"`
	TotalVolume       float64 `json:"total_volume"`
	PriceChangePct1h  float64 `json:"price_change_percentage_1h_in_currency"`
	PriceChangePct24h float64 `json:"price_change_percentage_24h"`
	PriceChangePct7d  float64 `json:"price_change_percentage_7d_in_currency"`
	LastUpdated       string  `json:"last_updated"`
	Brief             string  `json:"brief"`
}

func fetchMarkets(id, vs string) (*TokenMarketInfo, error) {
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/markets?vs_currency=%s&ids=%s&price_change_percentage=1h,24h,7d&sparkline=false",
		strings.ToLower(vs), strings.ToLower(id),
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "go-coingecko-client")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 兼容错误信息的直观提示
	if resp.StatusCode != http.StatusOK {
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return nil, fmt.Errorf("markets %d: %s", resp.StatusCode, body.Error)
	}

	var arr []TokenMarketInfo
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("no markets data for id=%s", id)
	}
	return &arr[0], nil
}

// 支持多个 id，返回 map[id]*TokenMarketInfo
func fetchMarketsMany(ids []string, vs string) (map[string]*TokenMarketInfo, error) {
	if len(ids) == 0 {
		return map[string]*TokenMarketInfo{}, nil
	}
	// 规范化
	for i := range ids {
		ids[i] = strings.ToLower(strings.TrimSpace(ids[i]))
	}
	vs = strings.ToLower(strings.TrimSpace(vs))

	// CoinGecko 单页最多 250；这里按传入数量设置 per_page
	perPage := len(ids)
	if perPage <= 0 {
		perPage = 1
	}
	if perPage > 250 {
		perPage = 250 // 如需 >250，自行拆批
	}

	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/markets?vs_currency=%s&ids=%s&price_change_percentage=1h,24h,7d&sparkline=false&per_page=%d&page=1",
		vs,
		strings.Join(ids, ","),
		perPage,
	)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "go-coingecko-client")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body.Error == "" {
			body.Error = resp.Status
		}
		return nil, fmt.Errorf("markets %d: %s", resp.StatusCode, body.Error)
	}

	var arr []TokenMarketInfo
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}

	out := make(map[string]*TokenMarketInfo, len(arr))
	for i := range arr {
		item := arr[i] // 取地址时要用下标变量
		out[item.ID] = &item
	}
	return out, nil
}

// ===== 派生指标（不依赖 OHLC）=====
type derived struct {
	IntradayMid     float64 // 当日中枢 (high+low)/2
	RoundSupport    float64 // 心理整数位支撑（百位向下）
	RoundResistance float64 // 心理整数位阻力（百位向上）
	PosToMid        float64 // 现价相对中枢的偏离
}

func deriveExtras(m *TokenMarketInfo) derived {
	base := 100.0
	mid := (m.High24h + m.Low24h) / 2
	return derived{
		IntradayMid:     mid,
		RoundSupport:    math.Floor(m.CurrentPrice/base) * base,
		RoundResistance: math.Ceil(m.CurrentPrice/base) * base,
		PosToMid:        m.CurrentPrice - mid,
	}
}

// ===== Prompt 构造 =====
func humanTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Local().Format("2006-01-02 15:04")
}

func buildPromptOneID(vs string, m *TokenMarketInfo, d derived) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s(%s) 行情快照 [%s]\n",
		strings.ToUpper(m.Symbol), m.Name, humanTime(m.LastUpdated))
	fmt.Fprintf(sb, "- 现价：%.2f %s\n- 当日高/低：%.2f / %.2f\n",
		m.CurrentPrice, strings.ToUpper(vs), m.High24h, m.Low24h)
	fmt.Fprintf(sb, "- 涨跌：1h %.2f%% | 24h %.2f%% | 7d %.2f%%\n",
		m.PriceChangePct1h, m.PriceChangePct24h, m.PriceChangePct7d)
	fmt.Fprintf(sb, "- 24h 成交额：$%.0f\n", m.TotalVolume)
	fmt.Fprintf(sb, "- 关键位：支撑 %.2f（整数位 %.0f）| 阻力 %.2f（整数位 %.0f）\n",
		m.Low24h, d.RoundSupport, m.High24h, d.RoundResistance)
	fmt.Fprintf(sb, "- 位置关系：现价相对中枢(%.2f) 为 %+.2f\n",
		d.IntradayMid, d.PosToMid)

	// 约束输出风格：专业、简短、四小时视角
	fmt.Fprintf(sb, `
请用专业分析师口吻，在不超过90字内总结“未来四小时”的简评，避免交易建议措辞， 简评中如果出现现价的价格值，用${price}代替`)
	return sb.String()
}

func GetTokenPricePrompt(id, vs string) (string, error) {
	//m, err := fetchMarkets(id, vs)
	m, err := Get().GetMarketInfo(id)
	if err != nil {
		return "", err
	}
	d := deriveExtras(m)
	prompt := buildPromptOneID(vs, m, d)
	return prompt, nil
}
