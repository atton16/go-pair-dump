package main

import (
	"context"
	"log"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/atton16/go-pair-dump/internal/app"
	"github.com/atton16/go-pair-dump/internal/services"
)

func main() {
	start := time.Now()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	// Get args
	var args = services.GetArgs()
	txt, _ := json.MarshalIndent(args, "", "  ")
	log.Printf("args: %s\n", txt)

	// Get config
	var config = services.GetConfig()
	txt, _ = json.MarshalIndent(config.Redact(), "", "  ")
	log.Printf("config: %s\n", txt)

	// Main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get redis
	var rd = services.GetRedis()
	if config.Notification.Enable {
		log.Println("notification: enabled")
		rd.Connect(ctx)
		defer rd.Close()
		app.NotifyOK(ctx, app.StateStart)
	} else {
		log.Println("notification: disabled")
	}

	// Mongo connect
	var mongoSvc = services.GetMongo()
	mongoSvc.Connect(ctx)
	defer mongoSvc.Disconnect(ctx)

	// Ensure index
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			primitive.E{Key: "symbol", Value: 1},
			primitive.E{Key: "interval", Value: 1},
			primitive.E{Key: "openTime", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(config.Mongo.KlinesIndexName),
	}
	log.Printf("ensureIndex: ensuring index %s...\n", config.Mongo.KlinesIndexName)
	indexCreated, err := app.EnsureIndex(ctx, config.Mongo.KlinesCollection, config.Mongo.KlinesIndexName, indexModel)
	if err != nil {
		app.NotifyError(ctx, app.ErrorEnsureIndex)
		log.Fatalf("error: %v", err)
	}
	if indexCreated == nil {
		log.Println("ensureIndex: index already exists, do nothing.")
	} else {
		log.Println("ensureIndex: index created!")
	}

	// Get symbols
	var symbols *[]string = app.GetSymbols(ctx)
	// log.Printf("symbols: %+v\n", symbols)
	log.Printf("total symbols: %d\n", len(*symbols))

	log.Printf("Start dumping klines for %d symbols, this might take a while...\n", len(*symbols))
	log.Printf("Progress report every %d seconds.", config.Pairdump.Progress.Interval)
	klinesCount := 0
	matchedCount := int64(0)
	upsertedCount := int64(0)

	progressCtx, progressCancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(config.Pairdump.Progress.Interval) * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("upsert: MatchedCount=%d, UpsertedCount=%d\n", matchedCount, upsertedCount)
			}
		}
	}(progressCtx)

	for _, symbol := range *symbols {
		var klines []services.BinanceKline = app.GetKlines(
			ctx,
			symbol,
			services.BinanceKlineInterval(config.Pairdump.Klines.Interval),
			config.Pairdump.Klines.Limit,
		)
		klines = app.KlinesWithoutUnclosedKline(klines)
		klinesCount += len(klines)
		// log.Printf("klines: %+v\n", klines)
		// log.Printf("total klines: %d\n", len(klines))
		var bulkWriteModels []mongo.WriteModel
		for _, kline := range klines {
			updateOne := mongo.NewUpdateOneModel()
			updateOne.SetFilter(bson.M{
				"symbol":   kline.Symbol,
				"interval": kline.Interval,
				"openTime": kline.OpenTime,
			})
			updateOne.SetUpdate(bson.M{
				"$setOnInsert": kline,
			})
			updateOne.SetUpsert(true)
			bulkWriteModels = append(bulkWriteModels, updateOne)
		}
		// BulkWrite
		result, err := mongoSvc.BulkWrite(ctx, config.Mongo.KlinesCollection, bulkWriteModels)
		if err != nil {
			app.NotifyError(ctx, app.ErrorBulkWrite)
			log.Fatalf("error: %v", err)
		}
		// log.Printf("%+v", result)
		matchedCount += result.MatchedCount
		upsertedCount += result.UpsertedCount
	}
	progressCancel()
	elapsed := time.Since(start)
	log.Printf("upsert: MatchedCount=%d, UpsertedCount=%d\n", matchedCount, upsertedCount)
	log.Printf("total klines: %d\n", klinesCount)
	app.NotifyOK(ctx, app.StateDone)
	log.Printf("Process took %s", elapsed)
	log.Println("Done")
}
