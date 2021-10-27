package app

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/atton16/go-pair-dump/internal/services"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSymbols(ctx context.Context) *[]string {
	var config = services.GetConfig()
	var binance = services.GetBinance()
	data, err := binance.ExchangeInfo()
	if err != nil {
		NotifyError(ctx, AppGetSymbols, err)
		log.Fatalf("error: %v", err)
	}
	// log.Printf("exchangeInfo: %+v\n", data)
	var filterPattern = regexp.MustCompile(config.Pairdump.FilterPattern)
	var symbols []string
	for _, symbol := range data.Symbols {
		if filterPattern.MatchString(symbol.Symbol) {
			symbols = append(symbols, symbol.Symbol)
		}
	}
	return &symbols
}

func GetKlines(ctx context.Context, symbol string, interval services.BinanceKlineInterval, limit int) []services.BinanceKline {
	var binance = services.GetBinance()
	opts := services.BinanceKlinesOptions{
		Limit: &limit,
	}
	data, err := binance.Klines(symbol, interval, &opts)
	if err != nil {
		NotifyError(ctx, AppGetKlines, err)
		log.Fatalf("error: %v", err)
	}
	return data
}

func KlinesWithoutUnclosedKline(klines []services.BinanceKline) []services.BinanceKline {
	outKlines := klines
	lastIndex := len(klines) - 1
	lastKline := klines[lastIndex]
	if time.Until(lastKline.CloseTime) > 0 {
		outKlines = outKlines[:lastIndex]
	}
	return outKlines
}

func EnsureIndex(ctx context.Context, col string, name string, model mongo.IndexModel) (*string, error) {
	var mongoSvc = services.GetMongo()

	indexes, err := mongoSvc.ListIndexes(ctx, col)
	if err != nil {
		return nil, err
	}
	indexFound := false
	for _, d := range indexes {
		n := d.Map()["name"]
		if n == name {
			indexFound = true
		}
	}

	if !indexFound {
		indexResult, err := mongoSvc.CreateIndex(ctx, col, model)
		if err != nil {
			return nil, err
		}
		return &indexResult, nil
	}
	return nil, nil
}
