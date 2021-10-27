# go-pair-dump

This program dump klines that passes `regexp` filter from Binance to MongoDB.

## Build

```bash
go build
```

## Usage

```bash
go-pair-dump -c ./configs/dev.yaml
```

## Example Output

Note that following sensitive config data will be `REDACTED`:

- Password part in mongo url
- Redis password

```bash
2021/10/27 14:50:55 args: {
  "Config": "./configs/dev.yaml"
}
2021/10/27 14:50:55 config: {
  "Pairdump": {
    "FilterPattern": "USDT$|USDC$|BUSD$|DAI$",
    "Klines": {
      "Interval": "1d",
      "Limit": 1000
    },
    "Progress": {
      "Interval": 30
    }
  },
  "Binance": {
    "ApiURL": "https://api.binance.com/"
  },
  "Mongo": {
    "URL": "mongodb://127.0.0.1:27017",
    "DB": "pairdump-test",
    "KlinesCollection": "klines",
    "KlinesIndexName": "symbol_interval_openTime"
  },
  "Notification": {
    "Enable": true,
    "RedisAddr": "127.0.0.1:6379",
    "RedisDB": 0,
    "RedisUsername": "",
    "RedisPassword": "",
    "Channel": "pairdump"
  }
}
2021/10/27 14:50:55 notification: enabled
2021/10/27 14:50:55 notification: NotifyOK -> start
2021/10/27 14:50:55 notification: NotifyOK -> result=0, error=<nil>
2021/10/27 14:50:55 ensureIndex: ensuring index symbol_interval_openTime...
2021/10/27 14:50:55 ensureIndex: index already exists, do nothing.
2021/10/27 14:50:58 total symbols: 657
2021/10/27 14:50:58 Start dumping klines for 657 symbols, this might take a while...
2021/10/27 14:50:58 Progress report every 30 seconds.
2021/10/27 14:51:28 upsert: MatchedCount=83788, UpsertedCount=81
2021/10/27 14:51:58 upsert: MatchedCount=162424, UpsertedCount=212
2021/10/27 14:52:28 upsert: MatchedCount=223711, UpsertedCount=374
2021/10/27 14:52:53 upsert: MatchedCount=249294, UpsertedCount=574
2021/10/27 14:52:53 total klines: 249868
2021/10/27 14:52:53 notification: NotifyOK -> done
2021/10/27 14:52:53 notification: NotifyOK -> result=0, error=<nil>
2021/10/27 14:52:53 Process took 1m57.459444279s
2021/10/27 14:52:53 Done
```

## License

Copyright 2021 Attawit Kittikrairit

Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
