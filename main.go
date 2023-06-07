package main

import (
	"agent/grpc"
	"agent/logger"
	"agent/plugin_manager"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"sync"
)

// Options 命令行参数
type Options struct {
	Configuration string `short:"c" long:"configuration" description:"the configuration file"`
}

var options Options

func main() {
	logger.InitLogger()
	// 加载命令行参数
	loadCommandLineParams()
	Reload()
}

// 加载命令行参数
func loadCommandLineParams() {
	_, err := flags.Parse(&options)
	if err != nil {
		logger.Logger.Error("Failed to load command line.", zap.String("errorMsg", err.Error()))
	}
	logger.Logger.Info("Successfully loaded command line arguments.",
		zap.Any("configuration", zap.String("configuration", options.Configuration)),
	)
}

var once sync.Once

func Reload() {
	// 加载配置文件
	config := loadConfig(&options)
	// 重载插件
	plugin_manager.Reload(config.Programs)
	// todo grpc目前先不考虑端口的改变，因为端口改变会导致所有长链接断开，
	// todo 服务端需要获取最新的端口进行重连，这里是否有意义，待考虑，所以暂时不进行实现
	once.Do(func() {
		grpc.StartGrpcServer(config.GrpcServerConfig)
	})
}
