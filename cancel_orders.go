//go:build ignore
// +build ignore

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const binanceBaseURL = "https://api.binance.com"

type BinanceClient struct {
	APIKey    string
	SecretKey string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	client := &BinanceClient{
		APIKey:    os.Getenv("BINANCE_API_KEY"),
		SecretKey: os.Getenv("BINANCE_SECRET_KEY"),
	}

	if client.APIKey == "" || client.SecretKey == "" {
		fmt.Println("❌ กรุณาตั้งค่า BINANCE_API_KEY และ BINANCE_SECRET_KEY")
		fmt.Println("💡 สร้างไฟล์ .env:")
		fmt.Println("   BINANCE_API_KEY=your_api_key")
		fmt.Println("   BINANCE_SECRET_KEY=your_secret_key")
		return
	}

	fmt.Println("🚨 Binance Grid Cancellation Tool")
	fmt.Println("⚠️  ยกเลิก Orders ทั้งหมดสำหรับ BOMEUSDT")
	fmt.Println("==========================================")

	// Cancel all BOMEUSDT orders
	symbol := "BOMEUSDT"
	fmt.Printf("🗑️ กำลังยกเลิกทุก orders สำหรับ %s...\n", symbol)

	if err := cancelAllOrdersStandalone(client, symbol); err != nil {
		log.Fatalf("❌ ไม่สามารถยกเลิก orders: %v", err)
	}

	fmt.Println("\n✅ ยกเลิก Grid สำเร็จแล้ว!")
	fmt.Println("💰 ตรวจสอบ Balance:")

	// Show current balances
	balances, err := getBalancesStandalone(client)
	if err != nil {
		fmt.Printf("❌ ไม่สามารถดึงยอดเงิน: %v\n", err)
		return
	}

	fmt.Printf("   USDT: %.2f\n", balances["USDT"])
	fmt.Printf("   BOME: %.0f\n", balances["BOME"])

	// Calculate portfolio value
	if balances["BOME"] > 0 {
		currentPrice, err := getCurrentPriceStandalone(client, "BOMEUSDT")
		if err == nil {
			bomeValue := balances["BOME"] * currentPrice
			totalValue := balances["USDT"] + bomeValue
			fmt.Printf("   BOME Value: $%.2f (@ $%.8f)\n", bomeValue, currentPrice)
			fmt.Printf("   Total Portfolio: $%.2f\n", totalValue)
		}
	}

	fmt.Println("\n🔄 Grid Trading ได้หยุดทำงานแล้ว")
}

// Cancel all orders - standalone version
func cancelAllOrdersStandalone(client *BinanceClient, symbol string) error {
	endpoint := "/api/v3/openOrders"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

	signature := createSignatureStandalone(params.Encode(), client.SecretKey)
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
		return fmt.Errorf("binance API error: %s", string(body))
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

		signature := createSignatureStandalone(params.Encode(), client.SecretKey)
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

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("🎯 ยกเลิกสำเร็จ %d/%d orders สำหรับ %s\n", canceledCount, len(orders), symbol)
	return nil
}

// Get balances - standalone version
func getBalancesStandalone(client *BinanceClient) (map[string]float64, error) {
	endpoint := "/api/v3/account"

	params := url.Values{}
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/1e6))

	signature := createSignatureStandalone(params.Encode(), client.SecretKey)
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

// Get current price - standalone version
func getCurrentPriceStandalone(client *BinanceClient, symbol string) (float64, error) {
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

// Create signature - standalone version
func createSignatureStandalone(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
