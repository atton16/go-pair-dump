package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var binanceOnce sync.Once
var myBinance *Binance

const (
	ExchangeInfoPath string = "/api/v3/exchangeInfo"
	KlinesPath       string = "/api/v3/klines"
)

type Binance struct {
	apiURL string
}

type BinanceKlineInterval string

const (
	OneMinute      BinanceKlineInterval = "1m"
	ThreeMinutes   BinanceKlineInterval = "3m"
	FiveMinutes    BinanceKlineInterval = "5m"
	FifteenMinutes BinanceKlineInterval = "15m"
	ThirtyMinutes  BinanceKlineInterval = "30m"
	OneHour        BinanceKlineInterval = "1h"
	TwoHours       BinanceKlineInterval = "2h"
	FourHours      BinanceKlineInterval = "4h"
	SixHours       BinanceKlineInterval = "6h"
	EightHours     BinanceKlineInterval = "8h"
	TwelveHours    BinanceKlineInterval = "12h"
	OneDay         BinanceKlineInterval = "1d"
	ThreeDays      BinanceKlineInterval = "3d"
	OneWeek        BinanceKlineInterval = "1w"
	OneMonth       BinanceKlineInterval = "1M"
)

type BinanceKlinesOptions struct {
	StartTime *int64
	EndTime   *int64
	Limit     *int
}

type BinanceExchangeInfo struct {
	Timezone   string `json:"timezone"`
	ServerTime int64  `json:"serverTime"`
	RateLimits []struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int32  `json:"intervalNum"`
		Limit         int32  `json:"limit"`
	} `json:"rateLimits"`
	ExchangeFilters []interface{} `json:"exchangeFilters"`
	Symbols         []struct {
		Symbol                     string   `json:"symbol"`
		Status                     string   `json:"status"`
		BaseAsset                  string   `json:"baseAsset"`
		BaseAssetPrecision         int8     `json:"baseAssetPrecision"`
		QuoteAsset                 string   `json:"quoteAsset"`
		QuotePrecision             int8     `json:"quotePrecision"`
		QuoteAssetPrecision        int8     `json:"quoteAssetPrecision"`
		BaseCommissionPrecision    int8     `json:"baseCommissionPrecision"`
		QuoteCommissionPrecision   int8     `json:"quoteCommissionPrecision"`
		OrderTypes                 []string `json:"orderTypes"`
		IcebergAllowed             bool     `json:"icebergAllowed"`
		OcoAllowed                 bool     `json:"ocoAllowed"`
		QuoteOrderQtyMarketAllowed bool     `json:"quoteOrderQtyMarketAllowed"`
		IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
		IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
		Filters                    []struct {
			FilterType string `json:"filterType"`
			// filterType: PRICE_FILTER
			MinPrice string `json:"minPrice,omitempty"`
			MaxPrice string `json:"maxPrice,omitempty"`
			TickSize string `json:"tickSize,omitempty"`
			// filterType: PERCENT_PRICE
			MultiplierUp   string `json:"multiplierUp,omitempty"`
			MultiplierDown string `json:"multiplierDown,omitempty"`
			// filterType: PERCENT_PRICE, MIN_NOTIONAL
			AvgPriceMins int8 `json:"avgPriceMins,omitempty"`
			// filterType: LOT_SIZE, MARKET_LOT_SIZE
			MinQty   string `json:"minQty,omitempty"`
			MaxQty   string `json:"maxQty,omitempty"`
			StepSize string `json:"stepSize,omitempty"`
			// filterType: MIN_NOTIONAL
			MinNotional   string `json:"minNotional,omitempty"`
			ApplyToMarket bool   `json:"applyToMarket,omitempty"`
			// filterType: ICEBERG_PARTS
			Limit int8 `json:"limit,omitempty"`
			// filterType: MAX_NUM_ORDERS
			MaxNumOrders int16 `json:"maxNumOrders,omitempty"`
			// filterType: MAX_NUM_ALGO_ORDERS
			MaxNumAlgoOrders int16 `json:"maxNumAlgoOrders,omitempty"`
		} `json:"filters"`
		Permissions []string `json:"permissions"`
	} `json:"symbols"`
}

type BinanceKline struct {
	Symbol                   string    `bson:"symbol"`
	Interval                 string    `bson:"interval"`
	OpenTime                 time.Time `bson:"openTime"`
	Open                     float64   `bson:"open"`
	High                     float64   `bson:"high"`
	Low                      float64   `bson:"low"`
	Close                    float64   `bson:"close"`
	Volume                   float64   `bson:"volume"`
	CloseTime                time.Time `bson:"closeTime"`
	QuoteAssetVolume         float64   `bson:"quoteAssetVolume"`
	NumberOfTrades           int64     `bson:"numberOfTrades"`
	TakerBuyBaseAssetVolume  float64   `bson:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume float64   `bson:"takerBuyQuoteAssetVolume"`
	CreatedAt                time.Time `bson:"createdAt"`
	UpdatedAt                time.Time `bson:"updatedAt"`
	// Ignore                   string    `bson:"ignore"`
}

func (k *BinanceKline) UnmarshalJSON(bs []byte) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	arr := []interface{}{}
	err := json.Unmarshal(bs, &arr)
	if err != nil {
		return err
	}
	k.OpenTime = time.UnixMilli(int64(arr[0].(float64)))
	k.Open, _ = strconv.ParseFloat(arr[1].(string), 64)
	k.High, _ = strconv.ParseFloat(arr[2].(string), 64)
	k.Low, _ = strconv.ParseFloat(arr[3].(string), 64)
	k.Close, _ = strconv.ParseFloat(arr[4].(string), 64)
	k.Volume, _ = strconv.ParseFloat(arr[5].(string), 64)
	k.CloseTime = time.UnixMilli(int64(arr[6].(float64)))
	k.QuoteAssetVolume, _ = strconv.ParseFloat(arr[7].(string), 64)
	k.NumberOfTrades = int64(arr[8].(float64))
	k.TakerBuyBaseAssetVolume, _ = strconv.ParseFloat(arr[9].(string), 64)
	k.TakerBuyQuoteAssetVolume, _ = strconv.ParseFloat(arr[10].(string), 64)
	now := time.Now()
	k.CreatedAt = now
	k.UpdatedAt = now
	// k.Ignore = arr[11].(string)
	return nil
}

func GetBinance() *Binance {
	binanceOnce.Do(func() {
		config := GetConfig()
		if _, err := url.Parse(config.Binance.ApiURL); err != nil {
			log.Fatalf("error: %v", err)
		}
		myBinance = &Binance{
			apiURL: config.Binance.ApiURL,
		}
	})
	return myBinance
}

func (b *Binance) getApiURL() *url.URL {
	u, _ := url.Parse(b.apiURL)
	return u
}

func (b *Binance) rateLimitWait(res *http.Response) bool {
	retryAfterHeader := res.Header["Retry-After"]
	if len(retryAfterHeader) > 0 {
		retryAfter, err := strconv.Atoi(retryAfterHeader[0])
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		log.Printf("Binance API rate limit reached, pause for %d seconds...\n", retryAfter)
		time.Sleep(time.Duration(retryAfter) * time.Second)
		log.Println("Resume")
		return true
	}
	return false
}

func (b *Binance) ExchangeInfo() (*BinanceExchangeInfo, error) {
	u := b.getApiURL()
	u.Path = path.Join(u.Path, ExchangeInfoPath)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if b.rateLimitWait(res) {
		return b.ExchangeInfo()
	}
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("%d:%s", res.StatusCode, body)
	}
	var data BinanceExchangeInfo
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (b *Binance) Klines(symbol string, interval BinanceKlineInterval, opts ...*BinanceKlinesOptions) ([]BinanceKline, error) {
	u := b.getApiURL()
	u.Path = path.Join(u.Path, KlinesPath)
	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("interval", string(interval))
	if len(opts) > 0 {
		opt := opts[0]
		if opt.StartTime != nil {
			q.Set("startTime", strconv.FormatInt(*opt.StartTime, 10))
		}
		if opt.EndTime != nil {
			q.Set("endTime", strconv.FormatInt(*opt.EndTime, 10))
		}
		if opt.Limit != nil {
			q.Set("limit", strconv.FormatInt(int64(*opt.Limit), 10))
		}
	}
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if b.rateLimitWait(res) {
		return b.Klines(symbol, interval, opts...)
	}
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("%d:%s", res.StatusCode, body)
	}
	data := []BinanceKline{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	for i := range data {
		data[i].Symbol = symbol
		data[i].Interval = string(interval)
	}
	return data, nil
}
