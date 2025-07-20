package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Scan for best coins based on new listings (≤30 days)
func scanBestCoins() ([]CoinInfo, error) {
	fmt.Println("🔍 กำลังค้นหาเหรียญใหม่ (≤30 วัน) ด้วยกระบวนการ 2 ขั้นตอน...")
	fmt.Println("📅 ขั้นตอน 1: ใช้ timeframe 3 เดือน (4 เดือนย้อนหลัง) กรองเหรียญใหม่")
	fmt.Println("📊 ขั้นตอน 2: ใช้ timeframe 1 วัน (144 วันย้อนหลัง) วิเคราะห์เหรียญที่ผ่านการกรอง")

	// Get 24hr ticker data
	fmt.Println("📈 กำลังดึงข้อมูลตลาด 24 ชั่วโมง...")
	tickers, err := get24hrTickers()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถดึงข้อมูลตลาด: %v", err)
	}

	// STEP 1: Filter new coins using monthly timeframe (4 months back)
	fmt.Println("🔍 STEP 1: กำลังกรองเหรียญใหม่ด้วย timeframe 3 เดือน...")
	var newCoinTickers []Ticker24hr

	fmt.Printf("📊 ตรวจสอบ %d สัญลักษณ์ด้วยข้อมูล 4 เดือนย้อนหลัง...\n", len(tickers))

	for i, ticker := range tickers {
		if i%100 == 0 {
			fmt.Printf("   กรองแล้ว %d/%d สัญลักษณ์...\n", i, len(tickers))
		}

		// Only check USDT pairs
		if !strings.HasSuffix(ticker.Symbol, "USDT") {
			continue
		}

		// Skip stablecoins and obvious old coins
		if isExcludedSymbol(ticker.Symbol) {
			continue
		}

		// Use monthly data (4 months) to filter new coins
		if isNewCoinMonthly(ticker.Symbol) {
			newCoinTickers = append(newCoinTickers, ticker)
		}
	}

	fmt.Printf("✅ STEP 1 เสร็จสิ้น: พบเหรียญใหม่ %d เหรียญ (จาก %d สัญลักษณ์)\n", len(newCoinTickers), len(tickers))

	if len(newCoinTickers) == 0 {
		return []CoinInfo{}, nil
	}

	// STEP 2: Analyze filtered coins with daily timeframe (144 days back)
	fmt.Println("🔍 STEP 2: วิเคราะห์เหรียญใหม่ด้วย timeframe 1 วัน (144 วันย้อนหลัง)...")

	// Define scan criteria for NEW coins
	criteria := ScanCriteria{
		MinVolume:       50000,    // 50K+ USDT volume (lower for newer coins)
		MaxPrice:        2.0,      // Maximum $2 (slightly higher for more options)
		MinPrice:        0.000001, // Minimum price
		MinPriceChange:  -90.0,    // Allow deep dips (new coins volatile)
		MaxPriceChange:  1000.0,   // Allow massive gains for new coins
		MinAge:          1,        // 1+ day minimum
		RequireRecovery: false,    // New coins don't need recovery history
		MinScore:        20.0,     // Lower threshold for more results
		MaxResults:      25,       // Show top 25 coins
	}

	var candidates []CoinInfo

	for i, ticker := range newCoinTickers {
		if i%5 == 0 {
			fmt.Printf("   วิเคราะห์แล้ว %d/%d เหรียญใหม่...\n", i, len(newCoinTickers))
		}

		// Use daily data (144 days) for detailed analysis
		coinInfo := processNewCoinTicker(ticker, criteria)
		if coinInfo != nil {
			candidates = append(candidates, *coinInfo)
		}
	}

	fmt.Printf("✅ STEP 2 เสร็จสิ้น: %d เหรียญผ่านเกณฑ์การวิเคราะห์\n", len(candidates))

	// Sort by score
	sortCoinsByScore(candidates)

	// Return top results
	maxResults := criteria.MaxResults
	if len(candidates) < maxResults {
		maxResults = len(candidates)
	}

	return candidates[:maxResults], nil
}

// analyzeCoinsForAccumulation analyzes coins with AI for accumulation opportunities
func analyzeCoinsForAccumulation(coins []CoinInfo) ([]AINewCoinAnalysis, error) {
	if len(coins) == 0 {
		return []AINewCoinAnalysis{}, nil
	}

	fmt.Printf("\n🤖 AI วิเคราะห์เหรียญใหม่สำหรับการสะสม...\n")

	// Call AI analysis
	analyses, err := analyzeNewCoinsWithAI(coins)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %v", err)
	}

	return analyses, nil // Return all analyses
}

// Process new coin ticker with detailed analysis
func processNewCoinTicker(ticker Ticker24hr, criteria ScanCriteria) *CoinInfo {
	// Parse numeric values
	price, err := strconv.ParseFloat(ticker.LastPrice, 64)
	if err != nil || price < criteria.MinPrice || price > criteria.MaxPrice {
		return nil
	}

	volume, err := strconv.ParseFloat(ticker.QuoteVolume, 64)
	if err != nil || volume < criteria.MinVolume {
		return nil
	}

	priceChange, err := strconv.ParseFloat(ticker.PriceChangePercent, 64)
	if err != nil || priceChange < criteria.MinPriceChange || priceChange > criteria.MaxPriceChange {
		return nil
	}

	// Calculate score focused on NEW coin potential
	score := calculateNewCoinScore(ticker, price, volume, priceChange)
	if score < criteria.MinScore {
		return nil
	}

	baseCoin := getBaseCoin(ticker.Symbol)

	return &CoinInfo{
		Symbol:      ticker.Symbol,
		BaseCoin:    baseCoin,
		Price:       price,
		Volume24h:   volume,
		PriceChange: priceChange,
		Score:       score,
		Reason:      generateNewCoinReason(ticker, price, volume, priceChange, score),
		AgeDays:     getCoinAgeDaysDetailed(ticker.Symbol),
		LastUpdated: time.Now(),
	}
}

// Sort coins by score (descending)
func sortCoinsByScore(coins []CoinInfo) {
	sort.Slice(coins, func(i, j int) bool {
		return coins[i].Score > coins[j].Score
	})
}

// Check if symbol should be excluded
func isExcludedSymbol(symbol string) bool {
	excluded := []string{
		// Stablecoins
		"BUSDUSDT", "USDCUSDT", "TUSDUSDT", "PAXUSDT", "DAIUSDT",
		// Major old coins
		"BTCUSDT", "ETHUSDT", "BNBUSDT", "XRPUSDT", "ADAUSDT",
		"DOGEUSDT", "SOLUSDT", "MATICUSDT", "DOTUSDT", "AVAXUSDT",
	}

	for _, exc := range excluded {
		if symbol == exc {
			return true
		}
	}
	return false
}

// Calculate score focused on NEW coin potential
func calculateNewCoinScore(ticker Ticker24hr, price, volume, priceChange float64) float64 {
	score := 0.0

	// Volume score (40% weight) - critical for new coins
	volumeScore := 0.0
	if volume >= 1000000 { // 1M+ exceptional for new coin
		volumeScore = 40.0
	} else if volume >= 500000 { // 500K+ very good
		volumeScore = 35.0
	} else if volume >= 200000 { // 200K+ good
		volumeScore = 30.0
	} else if volume >= 100000 { // 100K+ acceptable
		volumeScore = 25.0
	} else if volume >= 50000 { // 50K+ minimum
		volumeScore = 15.0
	}
	score += volumeScore

	// New coin price potential (30% weight) - very low price = high potential
	priceScore := 0.0
	if price <= 0.000001 {
		priceScore = 30.0 // Ultra micro cap new coin
	} else if price <= 0.00001 {
		priceScore = 28.0
	} else if price <= 0.0001 {
		priceScore = 25.0
	} else if price <= 0.001 {
		priceScore = 22.0
	} else if price <= 0.01 {
		priceScore = 18.0
	} else if price <= 0.1 {
		priceScore = 15.0
	} else if price <= 1.0 {
		priceScore = 10.0
	} else if price <= 2.0 {
		priceScore = 5.0
	}
	score += priceScore

	// New coin momentum (20% weight) - can be volatile
	momentumScore := 0.0
	if priceChange >= 50.0 { // Strong rally
		momentumScore = 20.0
	} else if priceChange >= 20.0 { // Good momentum
		momentumScore = 18.0
	} else if priceChange >= 0.0 { // Positive
		momentumScore = 15.0
	} else if priceChange >= -20.0 { // Small dip - good entry
		momentumScore = 18.0
	} else if priceChange >= -50.0 { // Bigger dip - opportunity
		momentumScore = 22.0
	} else {
		// Very deep dip - high risk/reward
		momentumScore = 15.0
	}
	score += momentumScore

	// Activity bonus (10% weight) - trade count important for new coins
	activityScore := 0.0
	if ticker.Count >= 50000 {
		activityScore = 10.0
	} else if ticker.Count >= 10000 {
		activityScore = 8.0
	} else if ticker.Count >= 5000 {
		activityScore = 6.0
	} else if ticker.Count >= 1000 {
		activityScore = 4.0
	}
	score += activityScore

	return score
}

// Generate reason for NEW coin selection (in Thai)
func generateNewCoinReason(ticker Ticker24hr, price, volume, priceChange, score float64) string {
	reasons := []string{}

	// Volume assessment for new coins
	if volume >= 500000 {
		reasons = append(reasons, "ปริมาณเทรดยอดเยี่ยมสำหรับเหรียญใหม่")
	} else if volume >= 100000 {
		reasons = append(reasons, "ปริมาณเทรดดีสำหรับการเข้าใหม่")
	}

	// Price potential
	if price <= 0.00001 {
		reasons = append(reasons, "ราคาเข้าต่ำมาก")
	} else if price <= 0.001 {
		reasons = append(reasons, "ราคาเข้าต่ำ")
	} else if price <= 0.1 {
		reasons = append(reasons, "โอกาสเข้าราคาต่ำ")
	}

	// Momentum for new coins
	if priceChange >= 20.0 {
		reasons = append(reasons, "โมเมนตัมขาขึ้นแรง")
	} else if priceChange >= -20.0 && priceChange < 0.0 {
		reasons = append(reasons, "โอกาสซื้อจังหวะดิป")
	} else if priceChange >= -50.0 && priceChange < -20.0 {
		reasons = append(reasons, "โอกาสเข้าจังหวะลดลึก")
	} else if priceChange >= 0.0 && priceChange < 20.0 {
		reasons = append(reasons, "การเข้าใหม่มีเสถียรภาพ")
	}

	// Score assessment
	if score >= 80.0 {
		reasons = append(reasons, "ศักยภาพเหรียญใหม่ยอดเยี่ยม")
	} else if score >= 60.0 {
		reasons = append(reasons, "ศักยภาพเหรียญใหม่สูง")
	} else if score >= 40.0 {
		reasons = append(reasons, "ศักยภาพเหรียญใหม่ดี")
	}

	return strings.Join(reasons, ", ")
}
