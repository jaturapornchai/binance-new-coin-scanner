# 🚀 Binance New Coin Scanner

AI-powered cryptocurrency scanner for detecting new coin listings on Binance with advanced accumulation analysis.

## ✨ Features

### 🎯 Core Functionality
- **2-Step Scanning Process**: Monthly (3-month) → Daily (144-day) timeframe analysis
- **AI-Powered Analysis**: Daily market structure analysis with accumulation signals  
- **Real-time Data**: Live data from Binance API
- **Thai Language Interface**: Complete Thai language support
- **JSON Output**: Structured data for integration

### 📊 Analysis Features
- **Market Structure Analysis**: Daily/Weekly support & resistance levels
- **Technical Indicators**: Moving averages (MA7, MA14, MA21, MA30)
- **Volume Profile**: 21-day volume analysis with trend detection
- **Risk Assessment**: Confidence levels and risk evaluation
- **Price Targets**: Accumulation zones, stop loss, and profit targets

### 🔍 New Coin Detection
- **Monthly Filter**: First-pass filtering using 3-month data
- **Daily Verification**: Detailed analysis with 144-day history
- **Age Calculation**: Accurate coin age determination
- **Pattern Recognition**: Advanced pattern-based estimation

## 🛠️ Installation

### Prerequisites
- Go 1.19 or higher
- Internet connection for Binance API access

### Quick Start
```bash
# Clone the repository
git clone https://github.com/yourusername/binance-new-coin-scanner.git
cd binance-new-coin-scanner

# Initialize Go module
go mod init binance-scanner
go mod tidy

# Run the scanner
go run .
```

## 🚦 Usage

### Basic Scan
```bash
go run .
```

### Sample Output
```
🚀 ตัวสแกนเหรียญใหม่ Binance
🆕 เฉพาะเหรียญใหม่ (≤30 วัน)
🎯 โอกาสเข้าก่อนใคร + AI วิเคราะห์การสะสม
===============================================

🔍 STEP 1: กำลังกรองเหรียญใหม่ด้วย timeframe 3 เดือน...
✅ STEP 1 เสร็จสิ้น: พบเหรียญใหม่ 16 เหรียญ (จาก 3183 สัญลักษณ์)

🔍 STEP 2: วิเคราะห์เหรียญใหม่ด้วย timeframe 1 วัน (144 วันย้อนหลัง)...
✅ STEP 2 เสร็จสิ้น: 8 เหรียญผ่านเกณฑ์การวิเคราะห์

🏆 เหรียญใหม่ยอดนิยมสำหรับการเข้าก่อนใคร:
อันดับ | สัญลักษณ์     | ราคา       | เปลี่ยน  | ปริมาณ    | คะแนน | วัน  
-------|---------------|------------|---------|-----------|-------|----- 
1       | SPKUSDT       | $0.03764300 |   -5.4% | $11164   K |  83.0 | 34   
2       | SAHARAUSDT    | $0.07885000 |   -0.3% | $6433    K |  83.0 | 25   

🤖 AI Analysis Results (JSON):
[
  {
    "symbol": "SPKUSDT",
    "shouldAccumulate": false,
    "reverseSignal": false,
    "confidence": "ต่ำ",
    "riskLevel": "สูง",
    "recommendedAction": "หลีกเลี่ยง",
    "technicalSummary": "เทรนด์ขาขึ้นระยะสั้น",
    "priceAction": "ย่านล่างของช่วงรายวัน"
  }
]
```

## 📈 How It Works

### Step 1: Monthly Filtering
- Analyzes 3-month timeframe data (4 months back)
- Filters ~3,000+ symbols down to potential new coins
- Uses volume and trading activity as primary filters

### Step 2: Daily Analysis
- Detailed analysis using 144-day daily data
- Applies advanced scoring algorithm
- Calculates accurate coin age
- Generates investment recommendations

### AI Analysis
- Daily market structure analysis
- Multi-timeframe trend detection (MA7, MA14, MA21, MA30)
- Volume profile analysis (21-day average)
- Support/Resistance level identification
- Risk assessment and confidence scoring

## 🎯 Selection Criteria

### New Coin Criteria
- **Age**: ≤30 days since listing
- **Volume**: 50K+ USDT daily minimum
- **Price Range**: $0.000001 - $2.00
- **Price Change**: -90% to +1000% (high volatility accepted)
- **Score**: Minimum 20+ points

### Scoring Algorithm
- **Volume Score (40%)**: Higher volume = higher score
- **Price Potential (30%)**: Lower price = higher potential
- **Momentum (20%)**: Price movement analysis
- **Activity (10%)**: Trade count and liquidity

## 🤖 AI Analysis Features

### Technical Analysis
- **Short-term trend**: MA7 vs MA14 comparison
- **Medium-term trend**: MA14 vs MA21 comparison
- **Long-term trend**: MA21 vs MA30 comparison
- **Daily structure**: 30-day high/low analysis
- **Weekly structure**: 7-day high/low analysis

### Decision Logic
- **High Confidence**: ≤15 days + daily support + volume + trend
- **Medium Confidence**: ≤30 days + trend + volume + no breakout
- **Wait Signal**: Daily breakout + volume + long trend
- **Accumulate**: Weekly support + volume

### Risk Assessment
- **Low Risk**: Strong trends with volume confirmation
- **Medium Risk**: Mixed signals or moderate confidence
- **High Risk**: Weak trends or low confidence

## 📊 Output Format

### Console Output
- Thai language interface with emoji indicators
- Tabulated results with key metrics
- Summary statistics and recommendations

### JSON Output
- Complete analysis data for integration
- All technical indicators and signals
- Risk metrics and price targets
- Timestamps for data freshness

## ⚠️ Disclaimer

This tool is for **educational and research purposes only**. 

- **Not Financial Advice**: All analysis results are algorithmic and should not be considered as investment advice
- **High Risk**: Cryptocurrency investments, especially new coins, carry high risk
- **Do Your Research**: Always conduct your own research before making investment decisions
- **No Guarantees**: Past performance does not guarantee future results

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

- Create an issue for bug reports
- Discussions for feature requests
- Wiki for additional documentation

## 🔗 Related Projects

- [Binance API Documentation](https://binance-docs.github.io/apidocs/)
- [Go Binance Client](https://github.com/adshao/go-binance)

---

**⚡ Made with Go | 🎯 Powered by Binance API | 🤖 Enhanced with AI**
