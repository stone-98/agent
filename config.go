package main

import (
	"agent/grpc"
	"agent/logger"
	"agent/service"
	"agent/util"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sync"
)

type Config struct {
	GrpcServerConfig grpc.ServerConfig          `mapstructure:"grpcServer"`
	Programs         []*program_service.Program `mapstructure:"program"`
	Lock             *sync.Mutex                `json:"-"`
	Md5              string                     `json:"-"`
}

var configInit sync.Once

var c *Config

func loadConfig(options *Options) *Config {
	viper.SetConfigFile(options.Configuration)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to read config file: %v", err))
	}
	// 初始化config
	configInit.Do(func() {
		c = &Config{Lock: &sync.Mutex{}}
	})
	// 加载config
	c.Lock.Lock()
	if err := viper.Unmarshal(c); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %s", err))
	}
	c.Lock.Unlock()
	c.printLatestConfig()
	// 监听配置
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		currentMd5, err := util.ReadFileMd5(options.Configuration)
		if err != nil {
			logger.Logger.Error("Failed to read configuration file.", zap.String("errorMsg", err.Error()))
		}
		if c.Md5 == currentMd5 {
			logger.Logger.Info("The configuration file has not changed, skip this processing.")
			return
		}
		c.Md5 = currentMd5
		logger.Logger.Info("Configuration file changes.")

		if err := viper.ReadInConfig(); err != nil {
			logger.Logger.Panic("Failed to read config file")
		}
		if err := viper.Unmarshal(&c); err != nil {
			logger.Logger.Panic("failed to reload config file")
		} else {
			logger.Logger.Info("Config reloaded successfully.")
			Reload()
		}
	})
	return c
}

func (c *Config) printLatestConfig() {
	jsonData, err := json.Marshal(c)
	if err != nil {
		logger.Logger.Error("Failed to marshal struct to JSON.", zap.String("errorMsg", err.Error()))
		return
	}
	logger.Logger.Info("Config reloaded successfully.", zap.String("configuration", string(jsonData)))
}
