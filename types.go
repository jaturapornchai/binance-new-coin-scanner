package main

import (
	"time"
)

// BinanceClient represents Binance API client
type BinanceClient struct {
	APIKey    string
	SecretKey string
}

// CoinInfo represents information about a scanned coin
type CoinInfo struct {
	Symbol      string
	BaseCoin    string
	Price       float64
	Volume24h   float64
	PriceChange float64
	Score       float64
	Reason      string
	AgeDays     int // จำนวนวันที่เข้า listing
	LastUpdated time.Time
}

// ScanCriteria defines criteria for coin scanning
type ScanCriteria struct {
	MinVolume       float64
	MaxPrice        float64
	MinPrice        float64
	MinPriceChange  float64
	MaxPriceChange  float64
	MinAge          int
	RequireRecovery bool
	MinScore        float64
	MaxResults      int
}

// Kline represents a candlestick
type Kline struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

// AIAnalysis represents the analysis result from AI
type AIAnalysis struct {
	Analysis          string
	Support           float64
	Resistance        float64
	BuyLevels         []float64
	SellLevels        []float64
	Confidence        string
	RiskLevel         string
	MaxPositionSize   string
	OptimalQuantity   float64
	RecommendedBudget string
	GapStrategy       string
}

// AINewCoinAnalysis represents AI analysis for new coins accumulation
type AINewCoinAnalysis struct {
	Symbol            string    `json:"symbol"`
	Price             float64   `json:"price"`
	AgeDays           int       `json:"ageDays"`
	ShouldAccumulate  bool      `json:"shouldAccumulate"`
	ReverseSignal     bool      `json:"reverseSignal"`
	Confidence        string    `json:"confidence"`        // "สูง", "ปานกลาง", "ต่ำ"
	RiskLevel         string    `json:"riskLevel"`         // "ต่ำ", "ปานกลาง", "สูง"
	RecommendedAction string    `json:"recommendedAction"` // "สะสม", "รอ", "หลีกเลี่ยง"
	AccumulationRange []float64 `json:"accumulationRange"` // [min_price, max_price]
	StopLoss          float64   `json:"stopLoss"`
	ProfitTarget      []float64 `json:"profitTarget"` // [target1, target2, target3]
	TechnicalSummary  string    `json:"technicalSummary"`
	MarketSentiment   string    `json:"marketSentiment"`
	VolumeAnalysis    string    `json:"volumeAnalysis"`
	PriceAction       string    `json:"priceAction"`
	TimeFrame         string    `json:"timeFrame"`
	LastUpdate        time.Time `json:"lastUpdate"`
}

// GridConfig represents grid trading configuration
type GridConfig struct {
	Symbol     string
	GridCount  int
	LowerPrice float64
	UpperPrice float64
	Investment float64
	CreatedAt  time.Time
}

// Ticker24hr represents 24hr ticker statistics
type Ticker24hr struct {
	Symbol             string
	PriceChange        string
	PriceChangePercent string
	WeightedAvgPrice   string
	PrevClosePrice     string
	LastPrice          string
	LastQty            string
	BidPrice           string
	BidQty             string
	AskPrice           string
	AskQty             string
	OpenPrice          string
	HighPrice          string
	LowPrice           string
	Volume             string
	QuoteVolume        string
	OpenTime           int64
	CloseTime          int64
	FirstId            int64
	LastId             int64
	Count              int64 `json:"count"`
}

// ExchangeInfo represents exchange information from Binance
type ExchangeInfo struct {
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	Symbols    []SymbolInfo `json:"symbols"`
}

// SymbolInfo represents symbol information including listing date
type SymbolInfo struct {
	Symbol               string `json:"symbol"`
	Status               string `json:"status"`
	BaseAsset            string `json:"baseAsset"`
	BaseAssetPrecision   int    `json:"baseAssetPrecision"`
	QuoteAsset           string `json:"quoteAsset"`
	QuoteAssetPrecision  int    `json:"quoteAssetPrecision"`
	OnboardDate          int64  `json:"onboardDate"`
	IsSpotTradingAllowed bool   `json:"isSpotTradingAllowed"`
}
