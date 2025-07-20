package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const binanceBaseURL = "https://api.binance.com"

// Place order on Binance
func placeOrder(client *BinanceClient, symbol, side, orderType, quantity, price string) (string, error) {
	endpoint := "/api/v3/order"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", side)
	params.Set("type", orderType)
	params.Set("quantity", quantity)
	params.Set("timeInForce", "GTC")
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

	if orderType == "LIMIT" {
		params.Set("price", price)
	}

	// Sign the request
	signature := createSignature(params.Encode(), client.SecretKey)
	params.Set("signature", signature)

	reqURL := binanceBaseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-MBX-APIKEY", client.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("binance API error: %s", string(body))
	}

	var orderResponse map[string]interface{}
	if err := json.Unmarshal(body, &orderResponse); err != nil {
		return "", err
	}

	orderID := fmt.Sprintf("%.0f", orderResponse["orderId"].(float64))
	return orderID, nil
}

// Get account balances
func getBalances(client *BinanceClient) (map[string]float64, error) {
	endpoint := "/api/v3/account"

	params := url.Values{}
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

	signature := createSignature(params.Encode(), client.SecretKey)
	params.Set("signature", signature)

	reqURL := binanceBaseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", client.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	var accountInfo struct {
		Balances []struct {
			Asset string `json:"asset"`
			Free  string `json:"free"`
		} `json:"balances"`
	}

	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, err
	}

	balances := make(map[string]float64)
	for _, balance := range accountInfo.Balances {
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			continue
		}
		if free > 0 {
			balances[balance.Asset] = free
		}
	}

	return balances, nil
}

// Get current price for specific symbol
func getCurrentPriceForSymbol(client *BinanceClient, symbol string) (float64, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", binanceBaseURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("binance API error: %s", string(body))
	}

	var priceResponse struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.Unmarshal(body, &priceResponse); err != nil {
		return 0, err
	}

	price, err := strconv.ParseFloat(priceResponse.Price, 64)
	if err != nil {
		return 0, err
	}

	return price, nil
}

// Get 24hr ticker statistics for all symbols
func get24hrTickers() ([]Ticker24hr, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/24hr", binanceBaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	var tickers []Ticker24hr
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, err
	}

	return tickers, nil
}

// Cancel all open orders for a symbol
func cancelAllOrders(client *BinanceClient, symbol string) error {
	// Get all open orders first
	endpoint := "/api/v3/openOrders"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

	signature := createSignature(params.Encode(), client.SecretKey)
	params.Set("signature", signature)

	reqURL := binanceBaseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-MBX-APIKEY", client.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("binance API error getting orders: %s", string(body))
	}

	var orders []map[string]interface{}
	if err := json.Unmarshal(body, &orders); err != nil {
		return err
	}

	if len(orders) == 0 {
		fmt.Printf("✅ ไม่มี orders ที่ต้องยกเลิกสำหรับ %s\n", symbol)
		return nil
	}

	fmt.Printf("🗑️ กำลังยกเลิก %d orders สำหรับ %s...\n", len(orders), symbol)

	// Cancel all orders
	canceledCount := 0
	for _, order := range orders {
		orderID := fmt.Sprintf("%.0f", order["orderId"].(float64))

		params := url.Values{}
		params.Set("symbol", symbol)
		params.Set("orderId", orderID)
		params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

		signature := createSignature(params.Encode(), client.SecretKey)
		params.Set("signature", signature)

		reqURL := binanceBaseURL + "/api/v3/order?" + params.Encode()

		req, err := http.NewRequest("DELETE", reqURL, nil)
		if err != nil {
			fmt.Printf("❌ ไม่สามารถสร้าง request ยกเลิก order %s: %v\n", orderID, err)
			continue
		}

		req.Header.Set("X-MBX-APIKEY", client.APIKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("❌ ไม่สามารถยกเลิก order %s: %v\n", orderID, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			canceledCount++
			fmt.Printf("✅ ยกเลิก order %s สำเร็จ\n", orderID)
		} else {
			fmt.Printf("❌ ไม่สามารถยกเลิก order %s (status: %d)\n", orderID, resp.StatusCode)
		}

		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	fmt.Printf("🎯 ยกเลิกสำเร็จ %d/%d orders สำหรับ %s\n", canceledCount, len(orders), symbol)
	return nil
}

// Get exchange info to check listing dates
func getExchangeInfo() (*ExchangeInfo, error) {
	url := binanceBaseURL + "/api/v3/exchangeInfo"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching exchange info: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var exchangeInfo ExchangeInfo
	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		return nil, fmt.Errorf("error unmarshaling exchange info: %v", err)
	}

	return &exchangeInfo, nil
}

// Check if symbol is a new coin using monthly timeframe (fast check first)
func isNewCoin(symbol string, exchangeInfo *ExchangeInfo) bool {
	// Create dummy client for kline requests (public endpoint)
	client := &BinanceClient{}

	// Step 1: Check monthly data first (fast check for coins ≤1 month old)
	monthlyKlines, err := getKlines(client, symbol, "1M", 2) // Get 2 months of data
	if err != nil {
		// If can't get monthly klines, assume it's too new
		return true
	}

	// Count months with actual trading volume
	activeMonths := 0
	for _, kline := range monthlyKlines {
		if kline.Volume > 0 {
			activeMonths++
		}
	}

	// If has 2+ months of active monthly data, it's NOT new
	if activeMonths >= 2 {
		return false
	}

	// If only 1 month or less, it's potentially new - confirm with daily data
	if activeMonths <= 1 {
		// Step 2: Use daily data to get exact age
		dailyKlines, err := getKlines(client, symbol, "1d", 35) // Get 35 days
		if err != nil {
			return true // Can't get daily data, assume new
		}

		// Count days with actual trading volume
		activeDays := 0
		for _, kline := range dailyKlines {
			if kline.Volume > 0 {
				activeDays++
			}
		}

		// If has ≤30 days of active trading, consider it new
		return activeDays <= 30
	}

	return false
}

// Check if coin is new using monthly timeframe (Step 1: 4 months history)
func isNewCoinMonthly(symbol string) bool {
	// Create dummy client for kline requests (public endpoint)
	client := &BinanceClient{}

	// Get 4 months of monthly data
	monthlyKlines, err := getKlines(client, symbol, "1M", 4)
	if err != nil {
		// If can't get monthly klines, assume it's too new
		return true
	}

	// Count months with actual trading volume
	activeMonths := 0
	for _, kline := range monthlyKlines {
		if kline.Volume > 0 {
			activeMonths++
		}
	}

	// If has 3+ months of active monthly data, it's NOT new
	// If has ≤2 months, consider it new
	return activeMonths <= 2
}

// Get accurate coin age using daily data (Step 2: 144 days history)
func getCoinAgeDaysDetailed(symbol string) int {
	// Create dummy client for kline requests (public endpoint)
	client := &BinanceClient{}

	// Get 144 days of daily data for detailed analysis
	dailyKlines, err := getKlines(client, symbol, "1d", 144)
	if err != nil {
		// Fallback to pattern-based estimation if API fails
		return estimateCoinAgeByPattern(symbol)
	}

	// Count days with actual trading volume
	activeDays := 0
	for _, kline := range dailyKlines {
		if kline.Volume > 0 {
			activeDays++
		}
	}

	if activeDays > 0 {
		return activeDays
	}

	// Fallback to pattern estimation
	return estimateCoinAgeByPattern(symbol)
}

// Get accurate coin age using daily data
func estimateCoinAgeDays(symbol string) int {
	// Create dummy client for kline requests (public endpoint)
	client := &BinanceClient{}

	// Get daily data for last 35 days
	dailyKlines, err := getKlines(client, symbol, "1d", 35)
	if err != nil {
		// Fallback to pattern-based estimation if API fails
		return estimateCoinAgeByPattern(symbol)
	}

	// Count days with actual trading volume
	activeDays := 0
	for _, kline := range dailyKlines {
		if kline.Volume > 0 {
			activeDays++
		}
	}

	if activeDays > 0 {
		return activeDays
	}

	// Fallback to pattern estimation
	return estimateCoinAgeByPattern(symbol)
}

// Fallback pattern-based age estimation
func estimateCoinAgeByPattern(symbol string) int {
	// Known very new coins (estimate)
	newCoins := map[string]int{
		"1000SATSUSDT": 15,
		"NEIROUSDT":    20,
		"DOGSUSDT":     25,
		"HMSTRUSDT":    18,
		"BOMEUSDT":     28,
		"PENGUUSDT":    12,
		"TURBOUSDT":    22,
		"RSRUSDT":      30,
		"CHESSUSDT":    35,
		"SPKUSDT":      14,
	}

	if age, exists := newCoins[symbol]; exists {
		return age
	}

	// Meme/trendy coins (likely newer)
	if strings.Contains(symbol, "MEME") || strings.Contains(symbol, "DOG") ||
		strings.Contains(symbol, "CAT") || strings.Contains(symbol, "PEPE") {
		return 45 // Estimate 45 days for meme coins
	}

	// Tokens with numbers (often newer)
	if strings.Contains(symbol, "1000") || strings.Contains(symbol, "1M") {
		return 25
	}

	// Default for others that passed new coin filter
	return 30 // Estimate 30 days for other "new" coins
}

// Create HMAC SHA256 signature
func createSignature(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// analyzeNewCoinsWithAI analyzes new coins with AI for accumulation signals
func analyzeNewCoinsWithAI(coins []CoinInfo) ([]AINewCoinAnalysis, error) {
	fmt.Println("🤖 กำลังวิเคราะห์เหรียญใหม่ด้วย AI...")

	var analyses []AINewCoinAnalysis

	// Create dummy client for public API calls (no auth needed for klines)
	client := &BinanceClient{}

	for i, coin := range coins {
		if i%3 == 0 {
			fmt.Printf("   AI วิเคราะห์แล้ว %d/%d เหรียญ...\n", i, len(coins))
		}

		// Get daily klines for detailed analysis (144 days)
		klines, err := getKlines(client, coin.Symbol, "1d", 144)
		if err != nil {
			fmt.Printf("⚠️ ไม่สามารถดึงข้อมูล %s: %v\n", coin.Symbol, err)
			continue
		}

		if len(klines) < 30 {
			fmt.Printf("⚠️ ข้อมูล %s ไม่เพียงพอสำหรับวิเคราะห์\n", coin.Symbol)
			continue
		}

		// Analyze with AI-like logic
		analysis := performAIAnalysis(coin, klines)
		analyses = append(analyses, analysis)
	}

	fmt.Printf("✅ AI วิเคราะห์เสร็จสิ้น: %d เหรียญ\n", len(analyses))
	return analyses, nil
}

// performAIAnalysis performs AI-like technical analysis focused on daily market structure
func performAIAnalysis(coin CoinInfo, klines []Kline) AINewCoinAnalysis {
	// Calculate technical indicators for daily timeframe analysis
	prices := make([]float64, len(klines))
	volumes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))

	for i, k := range klines {
		prices[i] = k.Close
		volumes[i] = k.Volume
		highs[i] = k.High
		lows[i] = k.Low
	}

	// Daily Market Structure Analysis
	// Short-term daily structure (7, 14, 21 days)
	ma7 := calculateSMA(prices, 7)
	ma14 := calculateSMA(prices, 14)
	ma21 := calculateSMA(prices, 21)

	// Medium-term daily structure (30 days for new coins)
	ma30 := calculateSMA(prices, 30)

	// Current market state based on daily structure
	currentPrice := prices[len(prices)-1]

	// Daily structure levels (30-day period for new coins)
	dailyPeriod := min(30, len(prices))
	recentDailyHigh := findMax(highs[len(highs)-dailyPeriod:])
	recentDailyLow := findMin(lows[len(lows)-dailyPeriod:])

	// Weekly structure levels (7-day period)
	weeklyPeriod := min(7, len(prices))
	weeklyHigh := findMax(highs[len(highs)-weeklyPeriod:])
	weeklyLow := findMin(lows[len(lows)-weeklyPeriod:])

	// Volume profile for daily analysis
	avgDailyVolume := calculateAverage(volumes[max(0, len(volumes)-21):]) // 21-day avg
	recentDailyVolume := volumes[len(volumes)-1]
	volumeTrend := recentDailyVolume > avgDailyVolume*1.2

	// Daily trend analysis
	shortTermTrend := len(ma7) > 0 && len(ma14) > 0 && ma7[len(ma7)-1] > ma14[len(ma14)-1]
	mediumTermTrend := len(ma14) > 0 && len(ma21) > 0 && ma14[len(ma14)-1] > ma21[len(ma21)-1]
	longTermTrend := len(ma21) > 0 && len(ma30) > 0 && ma21[len(ma21)-1] > ma30[len(ma30)-1]

	// Daily market structure breaks
	isDailyBreakout := currentPrice >= recentDailyHigh*0.98 // Near daily high breakout
	isDailySupport := currentPrice <= recentDailyLow*1.02   // Near daily low support
	isWeeklySupport := currentPrice <= weeklyLow*1.01       // Weekly low support

	// Enhanced AI Decision Logic based on daily market structure
	shouldAccumulate := false
	confidence := "ต่ำ"
	riskLevel := "สูง"
	recommendedAction := "หลีกเลี่ยง"

	// Daily structure-based accumulation logic for new coins
	if coin.AgeDays <= 15 && isDailySupport && volumeTrend && shortTermTrend {
		shouldAccumulate = true
		confidence = "สูง"
		riskLevel = "ปานกลาง"
		recommendedAction = "สะสม"
	} else if coin.AgeDays <= 30 && mediumTermTrend && volumeTrend && !isDailyBreakout {
		shouldAccumulate = true
		confidence = "ปานกลาง"
		riskLevel = "ปานกลาง"
		recommendedAction = "สะสม"
	} else if isDailyBreakout && volumeTrend && longTermTrend {
		shouldAccumulate = false
		confidence = "ปานกลาง"
		riskLevel = "ต่ำ"
		recommendedAction = "รอ"
	} else if isWeeklySupport && volumeTrend {
		shouldAccumulate = true
		confidence = "ปานกลาง"
		riskLevel = "ปานกลาง"
		recommendedAction = "สะสม"
	}

	// Enhanced reverse signal detection using daily structure
	reverseSignal := false
	if isDailySupport && coin.PriceChange < -10 && volumeTrend {
		reverseSignal = true
	} else if isWeeklySupport && coin.PriceChange < -15 && volumeTrend {
		reverseSignal = true
	}

	// Price targets based on daily structure levels
	accumulationRange := []float64{recentDailyLow * 0.95, recentDailyLow * 1.1}
	stopLoss := recentDailyLow * 0.85
	profitTarget := []float64{currentPrice * 1.3, currentPrice * 1.8, currentPrice * 2.5}

	// Generate summaries based on daily market structure
	technicalSummary := generateDailyTechnicalSummary(coin, shortTermTrend, mediumTermTrend, isDailySupport, volumeTrend)
	marketSentiment := generateMarketSentiment(coin.PriceChange, volumeTrend)
	volumeAnalysis := generateVolumeAnalysis(recentDailyVolume, avgDailyVolume)
	priceAction := generateDailyPriceAction(currentPrice, recentDailyHigh, recentDailyLow, weeklyHigh, weeklyLow)

	return AINewCoinAnalysis{
		Symbol:            coin.Symbol,
		Price:             currentPrice,
		AgeDays:           coin.AgeDays,
		ShouldAccumulate:  shouldAccumulate,
		ReverseSignal:     reverseSignal,
		Confidence:        confidence,
		RiskLevel:         riskLevel,
		RecommendedAction: recommendedAction,
		AccumulationRange: accumulationRange,
		StopLoss:          stopLoss,
		ProfitTarget:      profitTarget,
		TechnicalSummary:  technicalSummary,
		MarketSentiment:   marketSentiment,
		VolumeAnalysis:    volumeAnalysis,
		PriceAction:       priceAction,
		TimeFrame:         "1d",
		LastUpdate:        time.Now(),
	}
}

// Helper functions for technical analysis
func calculateSMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return []float64{}
	}

	sma := make([]float64, len(prices)-period+1)
	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += prices[j]
		}
		sma[i-period+1] = sum / float64(period)
	}
	return sma
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func findMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func findMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Generate daily technical summary based on market structure
func generateDailyTechnicalSummary(coin CoinInfo, shortTrend, mediumTrend, nearSupport, volumeTrend bool) string {
	summary := []string{}

	// Daily trend analysis
	if shortTrend && mediumTrend {
		summary = append(summary, "เทรนด์ขาขึ้นแข็งแกร่ง")
	} else if shortTrend {
		summary = append(summary, "เทรนด์ขาขึ้นระยะสั้น")
	} else if mediumTrend {
		summary = append(summary, "เทรนด์ขาขึ้นระยะกลาง")
	} else {
		summary = append(summary, "เทรนด์ขาลง")
	}

	// Daily support level analysis
	if nearSupport {
		summary = append(summary, "ใกล้แนวรับรายวัน")
	}

	// Volume confirmation
	if volumeTrend {
		summary = append(summary, "ปริมาณยืนยันการเคลื่อนไหว")
	}

	// Age-based analysis
	if coin.AgeDays <= 15 {
		summary = append(summary, "เหรียญใหม่มาก-โอกาสสูง")
	} else if coin.AgeDays <= 30 {
		summary = append(summary, "เหรียญใหม่-ศักยภาพดี")
	}

	return strings.Join(summary, ", ")
}

// Generate daily price action analysis
func generateDailyPriceAction(current, dailyHigh, dailyLow, weeklyHigh, weeklyLow float64) string {
	// Daily range analysis
	dailyRange := dailyHigh - dailyLow
	dailyPosition := (current - dailyLow) / dailyRange

	// Weekly range analysis
	weeklyRange := weeklyHigh - weeklyLow
	weeklyPosition := (current - weeklyLow) / weeklyRange

	if dailyPosition > 0.8 && weeklyPosition > 0.7 {
		return "ใกล้จุดสูงรายวันและรายสัปดาห์"
	} else if dailyPosition < 0.2 && weeklyPosition < 0.3 {
		return "ใกล้จุดต่ำรายวันและรายสัปดาห์"
	} else if dailyPosition > 0.6 {
		return "ย่านบนของช่วงรายวัน"
	} else if dailyPosition < 0.4 {
		return "ย่านล่างของช่วงรายวัน"
	} else if weeklyPosition > 0.6 {
		return "ย่านบนของช่วงรายสัปดาห์"
	} else if weeklyPosition < 0.4 {
		return "ย่านล่างของช่วงรายสัปดาห์"
	}
	return "ย่านกลางของโครงสร้างรายวัน"
}

// Generate analysis summaries
func generateTechnicalSummary(coin CoinInfo, trendingUp, nearSupport, volumeTrend bool) string {
	summary := []string{}

	if trendingUp {
		summary = append(summary, "เทรนด์ขาขึ้น")
	} else {
		summary = append(summary, "เทรนด์ขาลง")
	}

	if nearSupport {
		summary = append(summary, "ใกล้แนวรับ")
	}

	if volumeTrend {
		summary = append(summary, "ปริมาณเพิ่มขึ้น")
	}

	if coin.AgeDays <= 15 {
		summary = append(summary, "เหรียญใหม่มาก")
	} else if coin.AgeDays <= 30 {
		summary = append(summary, "เหรียญใหม่")
	}

	return strings.Join(summary, ", ")
}

func generateMarketSentiment(priceChange float64, volumeTrend bool) string {
	if priceChange > 10 && volumeTrend {
		return "แข็งแกร่งมาก"
	} else if priceChange > 0 && volumeTrend {
		return "บวกและมีปริมาณ"
	} else if priceChange < -20 {
		return "อ่อนแรงมาก"
	} else if priceChange < 0 {
		return "อ่อนแรง"
	}
	return "เป็นกลาง"
}

func generateVolumeAnalysis(current, average float64) string {
	ratio := current / average
	if ratio > 2.0 {
		return "ปริมาณสูงผิดปกติ"
	} else if ratio > 1.5 {
		return "ปริมาณสูงกว่าปกติ"
	} else if ratio > 1.2 {
		return "ปริมาณเพิ่มขึ้น"
	} else if ratio < 0.5 {
		return "ปริมาณต่ำมาก"
	}
	return "ปริมาณปกติ"
}

func generatePriceAction(current, recentHigh, recentLow float64) string {
	range_ := recentHigh - recentLow
	position := (current - recentLow) / range_

	if position > 0.8 {
		return "ใกล้แนวต้าน"
	} else if position < 0.2 {
		return "ใกล้แนวรับ"
	} else if position > 0.6 {
		return "ย่านบน"
	} else if position < 0.4 {
		return "ย่านล่าง"
	}
	return "ย่านกลาง"
}
