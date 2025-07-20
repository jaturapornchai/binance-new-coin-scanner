package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Analyze market with AI for dynamic symbol
func analyzeWithAI(client *BinanceClient, klines []Kline, usdtBalance float64, symbol string) (*AIAnalysis, error) {
	deepSeekKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepSeekKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY ไม่ได้ตั้งค่า")
	}

	klinesData := formatKlinesForAI(klines, symbol)

	// Get current price for context
	currentPrice, err := getCurrentPriceForSymbol(client, symbol)
	if err != nil {
		return nil, err
	}

	baseCoin := getBaseCoin(symbol)

	prompt := fmt.Sprintf(`%s

Current %s Price: $%.8f
Available USDT Balance: $%.2f

As a professional cryptocurrency trader and risk manager, analyze this comprehensive 36-hour %s data (144 × 15-minute candles) and provide EXACTLY 5 BUY levels and 5 SELL levels:

**REQUIREMENTS:**
- **Exactly 5 BUY orders**: Below current price for accumulation
- **Exactly 5 SELL orders**: Above current price for profit-taking
- **Optimal Gap Calculation**: Based on 36-hour volatility patterns and volume analysis
- **Fee Consideration**: Minimum 0.3%% profit gap to cover 0.2%% total fees + profit margin

**36-Hour Deep Analysis Focus:**
1. **Extended Volatility Pattern**: Calculate optimal spacing from 36-hour price movements
2. **Volume Accumulation Zones**: Place orders at significant volume clusters
3. **Multi-Day Support/Resistance**: Identify key S/R levels from 36-hour dataset
4. **Momentum Cycles**: Analyze 15-minute momentum shifts over 1.5 days
5. **Risk/Reward Optimization**: Balance fill probability with profit potential over extended timeframe

**%s 15m Trading Specifics:**
- 36-hour data reveals stronger support/resistance levels
- 15-minute timeframe reduces noise while capturing trends
- Extended analysis allows better gap optimization
- Volume patterns more reliable over longer periods
- Strategic order placement for accumulation opportunities

Please respond in JSON format with EXACTLY 5 buy levels and 5 sell levels:
{
  "analysis": "comprehensive 36-hour %s analysis focusing on optimal 5+5 grid placement using 15m timeframe",
  "support": strongest_support_from_36h_data,
  "resistance": strongest_resistance_from_36h_data,
  "buyLevels": [5_precise_buy_prices_below_current_price],
  "sellLevels": [5_precise_sell_prices_above_current_price],
  "confidence": "High/Medium/Low",
  "riskLevel": "Low/Medium/High", 
  "maxPositionSize": "percentage_string_like_20%%",
  "optimalQuantity": %s_quantity_per_order,
  "recommendedBudget": total_budget_based_on_36h_analysis,
  "gapStrategy": "explanation of 36-hour gap calculation and 15m timeframe reasoning for accumulation"
}

Focus on creating the most profitable 5+5 grid configuration using 36-hour 15-minute market intelligence for superior accumulation opportunities.`,
		klinesData, baseCoin, currentPrice, usdtBalance, baseCoin, baseCoin, symbol, strings.ToLower(baseCoin))

	requestBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
		"max_tokens":  1000,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deepSeekKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("DeepSeek API error: %s", string(body))
	}

	var aiResponse map[string]interface{}
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return nil, err
	}

	choices := aiResponse["choices"].([]interface{})
	if len(choices) == 0 {
		return nil, fmt.Errorf("ไม่มี response จาก AI")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	// Parse AI response
	analysis, err := parseAIResponse(content)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

// Parse AI response with flexible string/float handling
func parseAIResponse(content string) (*AIAnalysis, error) {
	var analysis AIAnalysis
	content = strings.TrimSpace(content)

	// Remove markdown code blocks
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	}
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	// Find JSON object boundaries
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}") + 1

	if start == -1 || end == 0 {
		return nil, fmt.Errorf("ไม่พบ JSON object ใน AI response")
	}

	jsonContent := content[start:end]

	// Parse JSON with flexible string/float handling
	var rawResponse map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &rawResponse); err != nil {
		return nil, fmt.Errorf("ไม่สามารถแยก JSON จาก AI response: %v\nContent: %s", err, jsonContent)
	}

	// Convert and validate data types
	analysis.Analysis, _ = rawResponse["analysis"].(string)
	analysis.Confidence, _ = rawResponse["confidence"].(string)
	analysis.RiskLevel, _ = rawResponse["riskLevel"].(string)
	analysis.MaxPositionSize, _ = rawResponse["maxPositionSize"].(string)
	analysis.RecommendedBudget, _ = rawResponse["recommendedBudget"].(string)
	analysis.GapStrategy, _ = rawResponse["gapStrategy"].(string)

	// Handle support/resistance (can be string or float)
	if supportStr, ok := rawResponse["support"].(string); ok {
		if support, err := strconv.ParseFloat(supportStr, 64); err == nil {
			analysis.Support = support
		}
	} else if support, ok := rawResponse["support"].(float64); ok {
		analysis.Support = support
	}

	if resistanceStr, ok := rawResponse["resistance"].(string); ok {
		if resistance, err := strconv.ParseFloat(resistanceStr, 64); err == nil {
			analysis.Resistance = resistance
		}
	} else if resistance, ok := rawResponse["resistance"].(float64); ok {
		analysis.Resistance = resistance
	}

	// Handle optimalQuantity
	if quantityFloat, ok := rawResponse["optimalQuantity"].(float64); ok {
		analysis.OptimalQuantity = quantityFloat
	}

	// Handle buyLevels array (can be strings or floats)
	if buyLevelsRaw, ok := rawResponse["buyLevels"].([]interface{}); ok {
		analysis.BuyLevels = make([]float64, 0, len(buyLevelsRaw))
		for _, level := range buyLevelsRaw {
			if levelStr, ok := level.(string); ok {
				if price, err := strconv.ParseFloat(levelStr, 64); err == nil {
					analysis.BuyLevels = append(analysis.BuyLevels, price)
				}
			} else if price, ok := level.(float64); ok {
				analysis.BuyLevels = append(analysis.BuyLevels, price)
			}
		}
	}

	// Handle sellLevels array (can be strings or floats)
	if sellLevelsRaw, ok := rawResponse["sellLevels"].([]interface{}); ok {
		analysis.SellLevels = make([]float64, 0, len(sellLevelsRaw))
		for _, level := range sellLevelsRaw {
			if levelStr, ok := level.(string); ok {
				if price, err := strconv.ParseFloat(levelStr, 64); err == nil {
					analysis.SellLevels = append(analysis.SellLevels, price)
				}
			} else if price, ok := level.(float64); ok {
				analysis.SellLevels = append(analysis.SellLevels, price)
			}
		}
	}

	return &analysis, nil
}

// Format klines data for AI analysis with dynamic symbol
func formatKlinesForAI(klines []Kline, symbol string) string {
	var result strings.Builder
	baseCoin := getBaseCoin(symbol)
	result.WriteString(fmt.Sprintf("15 Minute Candlestick Data for %s (Latest 36 hours - 144 candles):\n\n", symbol))

	// Show summary statistics first
	if len(klines) > 0 {
		high := klines[0].High
		low := klines[0].Low
		totalVolume := 0.0

		for _, k := range klines {
			if k.High > high {
				high = k.High
			}
			if k.Low < low {
				low = k.Low
			}
			totalVolume += k.Volume
		}

		result.WriteString("📊 36-Hour Summary:\n")
		result.WriteString(fmt.Sprintf("   High: $%.8f | Low: $%.8f | Range: %.2f%%\n",
			high, low, ((high-low)/low)*100))
		result.WriteString(fmt.Sprintf("   Total Volume: %.0f %s | Avg Volume: %.0f %s\n\n",
			totalVolume, baseCoin, totalVolume/float64(len(klines)), baseCoin))
	}

	// Show recent 16 candles (4 hours) with more detail for 15m timeframe
	start := len(klines) - 16
	if start < 0 {
		start = 0
	}

	result.WriteString("📈 Recent 4-Hour Detail (Latest 16 candles):\n")
	for i := start; i < len(klines); i++ {
		k := klines[i]
		timestamp := time.Unix(k.OpenTime/1000, 0)
		result.WriteString(fmt.Sprintf("Candle %d (%s): O:%.8f H:%.8f L:%.8f C:%.8f V:%.0f\n",
			i-start+1, timestamp.Format("Jan 2 15:04"), k.Open, k.High, k.Low, k.Close, k.Volume))
	}

	// Show 6-hour summary for 36-hour trend analysis (6 periods of 6 hours each)
	result.WriteString("\n📊 6-Hour Period Analysis (36 hours):\n")
	for period := 0; period < 6; period++ {
		startIdx := period * 24 // 24 candles = 6 hours for 15m
		endIdx := (period + 1) * 24
		if endIdx > len(klines) {
			endIdx = len(klines)
		}
		if startIdx >= len(klines) {
			break
		}

		periodOpen := klines[startIdx].Open
		periodClose := klines[endIdx-1].Close
		periodChange := ((periodClose - periodOpen) / periodOpen) * 100

		periodVolume := 0.0
		for i := startIdx; i < endIdx; i++ {
			periodVolume += klines[i].Volume
		}

		timestamp := time.Unix(klines[startIdx].OpenTime/1000, 0)
		result.WriteString(fmt.Sprintf("Period %d (%s): %.8f → %.8f (%+.2f%%) Vol: %.0f\n",
			period+1, timestamp.Format("Jan 2 15:04"), periodOpen, periodClose, periodChange, periodVolume))
	}

	return result.String()
}
