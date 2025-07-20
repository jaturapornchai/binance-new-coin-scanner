package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	fmt.Println("🚀 ตัวสแกนเหรียญใหม่ Binance")
	fmt.Println("🆕 เฉพาะเหรียญใหม่ (≤30 วัน)")
	fmt.Println("🎯 โอกาสเข้าก่อนใคร + AI วิเคราะห์การสะสม")
	fmt.Println("===============================================")

	// Scan for best coins
	fmt.Println("🔍 กำลังค้นหาเหรียญใหม่สำหรับการเข้าก่อนใคร...")
	bestCoins, err := scanBestCoins()
	if err != nil {
		log.Fatalf("❌ ไม่สามารถสแกนเหรียญได้: %v", err)
	}

	// Show results
	fmt.Printf("\n✅ พบเหรียญใหม่น่าสนใจ %d เหรียญ\n\n", len(bestCoins))

	if len(bestCoins) == 0 {
		fmt.Println("❌ ไม่พบเหรียญใหม่ที่ตรงตามเกณฑ์")
		fmt.Println("🔚 การวิเคราะห์เหรียญใหม่เสร็จสิ้น!")
		return
	}

	fmt.Println("🏆 เหรียญใหม่ยอดนิยมสำหรับการเข้าก่อนใคร:")
	fmt.Println("อันดับ | สัญลักษณ์     | ราคา       | เปลี่ยน  | ปริมาณ    | คะแนน | วัน  | ศักยภาพเหรียญใหม่")
	fmt.Println("-------|---------------|------------|---------|-----------|-------|------|------------------")

	for i, coin := range bestCoins {
		fmt.Printf("%-7d | %-13s | $%-9.8f | %+6.1f%% | $%-8.0fK | %5.1f | %-4d | %s\n",
			i+1,
			coin.Symbol,
			coin.Price,
			coin.PriceChange,
			coin.Volume24h/1000,
			coin.Score,
			coin.AgeDays,
			coin.Reason)
	}

	// AI Analysis for Accumulation
	fmt.Printf("\n🤖 AI วิเคราะห์การสะสมเหรียญใหม่...\n")
	aiAnalyses, err := analyzeCoinsForAccumulation(bestCoins)
	if err != nil {
		fmt.Printf("⚠️ AI analysis ล้มเหลว: %v\n", err)
	} else {
		// Convert to JSON
		jsonData, err := json.MarshalIndent(aiAnalyses, "", "  ")
		if err != nil {
			fmt.Printf("⚠️ ไม่สามารถแปลงเป็น JSON: %v\n", err)
		} else {
			fmt.Printf("\n📊 AI Analysis Results (JSON):\n")
			fmt.Println(string(jsonData))
		}

		// Show summary
		fmt.Printf("\n📋 สรุป AI Analysis:\n")
		shouldAccumulate := 0
		reverseSignals := 0
		for _, analysis := range aiAnalyses {
			if analysis.ShouldAccumulate {
				shouldAccumulate++
			}
			if analysis.ReverseSignal {
				reverseSignals++
			}
		}

		fmt.Printf("   • เหรียญที่วิเคราะห์: %d เหรียญ\n", len(aiAnalyses))
		fmt.Printf("   • แนะนำให้สะสม: %d เหรียญ\n", shouldAccumulate)
		fmt.Printf("   • มีสัญญาณกลับตัว: %d เหรียญ\n", reverseSignals)

		// Show top recommendations
		fmt.Printf("\n🎯 เหรียญแนะนำจาก AI:\n")
		for _, analysis := range aiAnalyses {
			if analysis.ShouldAccumulate || analysis.ReverseSignal {
				fmt.Printf("   • %s: %s (ความเชื่อมั่น: %s, ความเสี่ยง: %s)\n",
					analysis.Symbol,
					analysis.RecommendedAction,
					analysis.Confidence,
					analysis.RiskLevel)
				if analysis.ReverseSignal {
					fmt.Printf("     ⚡ มีสัญญาณกลับตัวขึ้น!\n")
				}
			}
		}
	}

	// Enhanced Summary for NEW Coin Analysis
	fmt.Printf("\n📊 สรุปการสแกนเหรียญใหม่:\n")
	fmt.Printf("   • เป้าหมาย: เฉพาะเหรียญใหม่ (≤30 วัน)\n")
	fmt.Printf("   • จำนวนสัญลักษณ์ที่วิเคราะห์: ~3,000+\n")
	fmt.Printf("   • เหรียญใหม่ที่ผ่านเกณฑ์: %d เหรียญ\n", len(bestCoins))
	fmt.Printf("   • คะแนนสูงสุด: %.1f\n", bestCoins[0].Score)
	fmt.Printf("   • เหรียญใหม่ยอดนิยม: %s\n", bestCoins[0].Symbol)

	fmt.Printf("\n💡 เกณฑ์การคัดเลือกเหรียญใหม่:\n")
	fmt.Printf("   • ปริมาณขั้นต่ำ: 50K+ USDT ต่อวัน\n")
	fmt.Printf("   • ช่วงราคา: $0.000001 - $2 (ราคาต่ำ)\n")
	fmt.Printf("   • ช่วงการเปลี่ยนแปลง: -90%% ถึง +1000%% (ความผันผวนสูง)\n")
	fmt.Printf("   • โฟกัส: เหรียญที่เข้าใหม่ (ไม่รวมเหรียญใหญ่เก่า)\n")
	fmt.Printf("   • คะแนนขั้นต่ำ: 20+ คะแนน\n")

	fmt.Printf("\n🎯 แนะนำสำหรับการเข้าก่อนใคร:\n")
	fmt.Printf("   สัญลักษณ์หลัก: %s\n", bestCoins[0].Symbol)
	fmt.Printf("   ราคาเข้า: $%.8f\n", bestCoins[0].Price)
	fmt.Printf("   กลยุทธ์: %s\n", bestCoins[0].Reason)
	fmt.Printf("   ระยะเวลา: ช่วงการสะสมก่อนใคร\n")

	fmt.Println("\n🔚 การวิเคราะห์เหรียญใหม่ + AI Analysis เสร็จสิ้น!")
}
