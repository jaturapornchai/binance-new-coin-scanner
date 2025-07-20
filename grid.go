package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// Place grid orders with dynamic symbol support
func placeGridOrders(client *BinanceClient, analysis *AIAnalysis, symbol string, usdtBalance float64) error {
	baseCoin := getBaseCoin(symbol)

	// Calculate available budget (use 80% of balance for safety)
	availableBudget := usdtBalance * 0.8

	// Calculate quantity per buy order
	quantityPerOrder := availableBudget / 5.0 // 5 buy orders

	fmt.Printf("\n🎯 Placing %s Grid Orders:\n", symbol)
	fmt.Printf("   Available Budget: $%.2f USDT\n", availableBudget)
	fmt.Printf("   Per Order Budget: $%.2f USDT\n", quantityPerOrder)

	// Place buy orders
	fmt.Printf("\n📉 Placing 5 Buy Orders for %s:\n", symbol)
	buyOrdersPlaced := 0
	for i, price := range analysis.BuyLevels {
		if price <= 0 {
			continue
		}

		// Calculate quantity based on price and budget
		quantity := quantityPerOrder / price

		// Format quantity according to symbol's precision
		formattedQuantity := formatQuantityForSymbol(quantity, symbol)

		if formattedQuantity <= 0 {
			fmt.Printf("   Buy %d: Skipped (quantity too small)\n", i+1)
			continue
		}

		fmt.Printf("   Buy %d: %.0f %s @ $%.8f (Total: $%.2f)\n",
			i+1, formattedQuantity, baseCoin, price, formattedQuantity*price)

		orderId, err := placeOrder(client, symbol, "BUY", "LIMIT",
			strconv.FormatFloat(formattedQuantity, 'f', -1, 64),
			strconv.FormatFloat(price, 'f', -1, 64))

		if err != nil {
			fmt.Printf("   ❌ Buy order %d failed: %v\n", i+1, err)
			continue
		}

		fmt.Printf("   ✅ Buy order %d placed! Order ID: %s\n", i+1, orderId)
		buyOrdersPlaced++
		time.Sleep(200 * time.Millisecond)
	}

	// Get current balance after buy orders for sell calculations
	balances, err := getBalances(client)
	if err != nil {
		return fmt.Errorf("ไม่สามารถดึงยอดเงินหลังวางคำสั่งซื้อ: %v", err)
	}

	// Extract coin balance for sell orders
	coinBalance := extractCoinBalance(balances, baseCoin)

	// Place sell orders
	fmt.Printf("\n📈 Placing 5 Sell Orders for %s:\n", symbol)
	sellOrdersPlaced := 0

	if coinBalance > 0 {
		// Use existing balance for sell orders
		quantityPerSell := coinBalance / 5.0

		for i, price := range analysis.SellLevels {
			if price <= 0 {
				continue
			}

			formattedQuantity := formatQuantityForSymbol(quantityPerSell, symbol)

			if formattedQuantity <= 0 {
				fmt.Printf("   Sell %d: Skipped (quantity too small)\n", i+1)
				continue
			}

			fmt.Printf("   Sell %d: %.0f %s @ $%.8f (Total: $%.2f)\n",
				i+1, formattedQuantity, baseCoin, price, formattedQuantity*price)

			orderId, err := placeOrder(client, symbol, "SELL", "LIMIT",
				strconv.FormatFloat(formattedQuantity, 'f', -1, 64),
				strconv.FormatFloat(price, 'f', -1, 64))

			if err != nil {
				fmt.Printf("   ❌ Sell order %d failed: %v\n", i+1, err)
				continue
			}

			fmt.Printf("   ✅ Sell order %d placed! Order ID: %s\n", i+1, orderId)
			sellOrdersPlaced++
			time.Sleep(200 * time.Millisecond)
		}
	} else {
		// No coin balance - can only place conditional sell orders based on expected buy fills
		expectedQuantityPerBuy := quantityPerOrder / ((analysis.BuyLevels[0] + analysis.BuyLevels[len(analysis.BuyLevels)-1]) / 2.0)
		expectedCoinBalance := expectedQuantityPerBuy * float64(buyOrdersPlaced)
		quantityPerSell := expectedCoinBalance / 5.0

		for i, price := range analysis.SellLevels {
			if price <= 0 {
				continue
			}

			formattedQuantity := formatQuantityForSymbol(quantityPerSell, symbol)

			if formattedQuantity <= 0 {
				fmt.Printf("   Sell %d: Skipped (quantity too small)\n", i+1)
				continue
			}

			fmt.Printf("   Sell %d: %.0f %s @ $%.8f (Conditional - pending buy fills)\n",
				i+1, formattedQuantity, baseCoin, price)
			sellOrdersPlaced++
		}
	}

	// Summary
	fmt.Printf("\n📊 %s Grid Summary:\n", symbol)
	fmt.Printf("   Buy Orders Placed: %d/5\n", buyOrdersPlaced)
	fmt.Printf("   Sell Orders Placed: %d/5\n", sellOrdersPlaced)
	fmt.Printf("   Total Orders: %d/10\n", buyOrdersPlaced+sellOrdersPlaced)

	if buyOrdersPlaced == 0 && sellOrdersPlaced == 0 {
		return fmt.Errorf("ไม่สามารถวางคำสั่งใดๆ ได้")
	}

	return nil
}

// Format quantity according to symbol's precision requirements
func formatQuantityForSymbol(quantity float64, symbol string) float64 {
	// Get base coin for precision settings
	baseCoin := getBaseCoin(symbol)

	// Default precision rules - could be enhanced with API data
	switch baseCoin {
	case "BTC", "ETH", "BNB":
		return roundToDecimals(quantity, 6)
	case "ADA", "DOT", "LINK", "SOL":
		return roundToDecimals(quantity, 2)
	case "DOGE", "SHIB", "BONK":
		return roundToDecimals(quantity, 0) // Whole numbers
	default:
		// For unknown coins, use reasonable precision
		if quantity >= 1000 {
			return roundToDecimals(quantity, 0)
		} else if quantity >= 10 {
			return roundToDecimals(quantity, 1)
		} else {
			return roundToDecimals(quantity, 2)
		}
	}
}

// Round to specified decimal places
func roundToDecimals(value float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return float64(int(value*multiplier+0.5)) / multiplier
}

// Extract coin balance from account balances
func extractCoinBalance(balances map[string]float64, coin string) float64 {
	if balance, exists := balances[coin]; exists {
		return balance
	}
	return 0.0
}

// Monitor grid performance for dynamic symbol
func monitorGridPerformance(client *BinanceClient, symbol string) error {
	fmt.Printf("\n👀 Starting %s Grid Performance Monitor...\n", symbol)

	tickerCount := 0

	for {
		time.Sleep(60 * time.Second) // Check every minute
		tickerCount++

		// Get current price
		currentPrice, err := getCurrentPriceForSymbol(client, symbol)
		if err != nil {
			log.Printf("❌ Error getting %s price: %v", symbol, err)
			continue
		}

		// Get account balances
		balances, err := getBalances(client)
		if err != nil {
			log.Printf("❌ Error getting balances: %v", err)
			continue
		}

		usdtBalance := balances["USDT"]
		baseCoin := getBaseCoin(symbol)
		coinBalance := extractCoinBalance(balances, baseCoin)

		// Calculate portfolio value in USDT
		portfolioValue := usdtBalance + (coinBalance * currentPrice)

		fmt.Printf("\n⏰ %s Monitor #%d:\n", symbol, tickerCount)
		fmt.Printf("   Current Price: $%.8f\n", currentPrice)
		fmt.Printf("   USDT Balance: $%.2f\n", usdtBalance)
		fmt.Printf("   %s Balance: %.2f (≈ $%.2f)\n", baseCoin, coinBalance, coinBalance*currentPrice)
		fmt.Printf("   Portfolio Value: $%.2f\n", portfolioValue)

		// Every 10 minutes, show more detailed status
		if tickerCount%10 == 0 {
			fmt.Printf("\n📊 %s Detailed Status (10-min update):\n", symbol)

			// Could add order status checking here
			// orders, err := getOpenOrders(client, symbol)
			// if err == nil {
			//     fmt.Printf("   Open Orders: %d\n", len(orders))
			// }
		}
	}
}

// Check if an order was filled and rebalance if needed
func checkAndRebalanceGrid(client *BinanceClient, symbol string, analysis *AIAnalysis) error {
	// This function could monitor filled orders and replace them
	// For now, just return nil - can be enhanced later
	return nil
}
