package services

import (
	"io/ioutil"
	"log"
	"net/url"
	"sync"

	"gopkg.in/yaml.v2"
)

var configOnce sync.Once
var myConfig *Config

type Config struct {
	Binance struct {
		ApiURL        string `yaml:"apiURL"`
		FilterPattern string `yaml:"filterPattern"`
		Klines        struct {
			Interval string `yaml:"interval"`
			Limit    int    `yaml:"limit"`
		} `yaml:"klines"`
		Progress struct {
			Interval int64 `yaml:"interval"`
		} `yaml:"progress"`
	} `yaml:"binance"`
	Mongo struct {
		URL     string `yaml:"url"`
		DB      string `yaml:"db"`
		Binance struct {
			SymbolsCollection string `yaml:"symbolsCollection"`
			SymbolsIndexName  string `yaml:"symbolsIndexName"`
			KlinesCollection  string `yaml:"klinesCollection"`
			KlinesIndexName   string `yaml:"klinesIndexName"`
		} `yaml:"binance"`
	} `yaml:"mongo"`
	Notification struct {
		Enable        bool   `yaml:"enable"`
		RedisAddr     string `yaml:"redisAddr"`
		RedisDB       int    `yaml:"redisDB"`
		RedisUsername string `yaml:"redisUsername"`
		RedisPassword string `yaml:"redisPassword"`
		Channel       string `yaml:"channel"`
	} `yaml:"notification"`
}

func GetConfig() *Config {
	configOnce.Do(func() {
		args := GetArgs()
		content, err := ioutil.ReadFile(args.Config)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		var config Config
		err = yaml.Unmarshal(content, &config)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		myConfig = &config
	})
	return myConfig
}

func (c *Config) Redact() Config {
	var copyOf = *c
	u, _ := url.Parse(copyOf.Mongo.URL)
	if u.User != nil {
		p, _ := u.User.Password()
		if p != "" {
			u.User = url.UserPassword(u.User.Username(), "REDACTED")
		}
		copyOf.Mongo.URL = u.String()
	}
	if copyOf.Notification.RedisPassword != "" {
		copyOf.Notification.RedisPassword = "REDACTED"
	}
	return copyOf
}
