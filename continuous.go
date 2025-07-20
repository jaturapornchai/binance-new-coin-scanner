package main

import (
	"fmt"
	"time"
)

// Run continuous grid bot for selected symbol
func runContinuousBot(client *BinanceClient, symbol string) {
	fmt.Printf("🔄 เริ่ม Continuous Grid Bot สำหรับ %s\n", symbol)
	fmt.Println("⏰ ทำงานทุก 15 นาที | กด Ctrl+C เพื่อหยุด")

	// Run immediately first time
	runSingleIteration(client, symbol, 1)

	// Create ticker for 15-minute intervals
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	iteration := 2

	for {
		select {
		case <-ticker.C:
			fmt.Printf("\n⏰ === รอบที่ %d - %s ===\n", iteration, time.Now().Format("15:04:05"))
			runSingleIteration(client, symbol, iteration)
			iteration++
		}
	}
}

// Run single trading iteration for specific symbol
func runSingleIteration(client *BinanceClient, symbol string, iteration int) {
	fmt.Printf("\n🚀 เริ่มรอบที่ %d สำหรับ %s\n", iteration, symbol)

	// Step 1: Cancel existing orders (optional - implement if needed)
	fmt.Println("🔄 ขั้นตอน 1: ยกเลิก orders เก่าทั้งหมด")
	// if err := cancelAllOrders(client, symbol); err != nil {
	//     fmt.Printf("❌ Error canceling orders: %v\n", err)
	//     return
	// }

	// Step 2: Get balance
	fmt.Println("💰 ขั้นตอน 2: ตรวจสอบยอดเงิน")
	balances, err := getBalances(client)
	if err != nil {
		fmt.Printf("❌ Error getting balances: %v\n", err)
		return
	}

	// Extract balances dynamically
	usdtBalance, coinBalance := extractBalances(balances, symbol)

	fmt.Printf("💵 USDT Balance: $%.2f\n", usdtBalance)
	fmt.Printf("💎 %s Balance: %.4f\n", getBaseCoin(symbol), coinBalance)

	if usdtBalance < 5 && coinBalance < 0.1 {
		fmt.Printf("❌ ไม่มียอดเงินเพียงพอ: USDT $%.2f, %s %.4f\n",
			usdtBalance, getBaseCoin(symbol), coinBalance)
		return
	}

	// Step 3: Get current price
	fmt.Println("💰 ขั้นตอน 3: ตรวจสอบราคาปัจจุบัน")
	currentPrice, err := getCurrentPriceForSymbol(client, symbol)
	if err != nil {
		fmt.Printf("❌ Error getting price: %v\n", err)
		return
	}
	fmt.Printf("💰 ราคา %s: $%.8f\n", symbol, currentPrice)

	// Step 4: Get 15m klines data (144 candles = 36 hours)
	fmt.Println("📊 ขั้นตอน 4: ดึงข้อมูล 15m Candlestick (36 ชั่วโมงย้อนหลัง)")
	klines, err := getKlines(client, symbol, "15m", 144)
	if err != nil {
		fmt.Printf("❌ ไม่สามารถดึงข้อมูล klines: %v\n", err)
		return
	}
	fmt.Printf("✅ ดึงข้อมูล %d candles (15m timeframe, 36h history)\n", len(klines))

	// Step 5: AI Analysis with 36-hour 15m data
	fmt.Println("🤖 ขั้นตอน 5: วิเคราะห์ด้วย AI (15m Deep Analysis)")
	analysis, err := analyzeWithAI(client, klines, usdtBalance, symbol)
	if err != nil {
		fmt.Printf("❌ ไม่สามารถวิเคราะห์ด้วย AI: %v\n", err)
		return
	}

	fmt.Printf("✅ AI Analysis สำเร็จ!\n")
	fmt.Printf("📝 วิเคราะห์ (5+5): %s\n", analysis.Analysis)
	fmt.Printf("🔻 Support: $%.8f\n", analysis.Support)
	fmt.Printf("📈 Resistance: $%.8f\n", analysis.Resistance)
	fmt.Printf("📊 BUY Levels: %v\n", analysis.BuyLevels)
	fmt.Printf("📊 SELL Levels: %v\n", analysis.SellLevels)
	fmt.Printf("🎯 Confidence: %s\n", analysis.Confidence)
	fmt.Printf("⚠️ Risk Level: %s\n", analysis.RiskLevel)
	fmt.Printf("💰 AI Budget: %s\n", analysis.RecommendedBudget)
	fmt.Printf("🧠 Gap Strategy: %s\n", analysis.GapStrategy)

	// Step 6: Create Grid Configuration (embedded)
	fmt.Println("⚙️ ขั้นตอน 6: สร้าง Grid Configuration")
	// Grid config is embedded in analysis

	// Step 7: Place orders
	fmt.Println("🎯 ขั้นตอน 7: วาง AI-Optimized Grid Orders")
	if err := placeGridOrders(client, analysis, symbol, usdtBalance); err != nil {
		fmt.Printf("❌ ไม่สามารถวาง Grid Orders: %v\n", err)
		return
	}

	fmt.Printf("✅ รอบที่ %d เสร็จสิ้น - รอ 15 นาทีถัดไป\n", iteration)
	fmt.Printf("📈 Trading Range: $%.8f - $%.8f\n", analysis.Support, analysis.Resistance)
	fmt.Printf("💰 ใช้งบ: %s (AI 5+5 15m Strategy สำหรับ %s)\n", analysis.RecommendedBudget, symbol)
}

// Extract USDT and coin balance dynamically based on symbol
func extractBalances(balances map[string]float64, symbol string) (float64, float64) {
	usdtBalance := balances["USDT"]
	baseCoin := getBaseCoin(symbol)
	coinBalance := balances[baseCoin]
	return usdtBalance, coinBalance
}

// Get base coin name from trading pair symbol
func getBaseCoin(symbol string) string {
	// Remove USDT suffix to get base coin
	if len(symbol) > 4 && symbol[len(symbol)-4:] == "USDT" {
		return symbol[:len(symbol)-4]
	}
	return symbol
}
