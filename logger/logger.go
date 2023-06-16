package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path/filepath"
)

var Logger *zap.Logger

func InitLogger() {
	// 1 用new自定义log日志
	// zap.New(xxx)
	// 2 zap.New需要接收一个core，core是zapcore.Core类型，zapcore.Core是一个interface类型，
	//   而zapcore.NewCore返回的ioCore刚好实现了这个接口类型的所有5个方法，那么NewCore也可以认为是core类型
	// 3 所以zap.New(core)变成了zap.New(zapcore.NewCore)
	// 4 而zapcore.NewCore需要三个变量：Encoder, WriteSyncer, LevelEnabler,我们在创建NewCore时自定义这三个类型变量即可，其中：
	//         Encoder：编码器 (写入日志格式)
	//         WriteSyncer：指定日志写到哪里去
	//         LevelEnabler：日志打印级别
	// NewCore(enc Encoder, ws WriteSyncer, enab LevelEnabler)

	// 4.2 通过GetEncoder获取自定义的Encoder
	encoder := getEncoder()
	// 4.4 通过GetWriteSyncer获取自定义的WriteSyncer
	fileSyncer := getFileSyncer()
	// 创建控制台输出
	consoleSyncer := zapcore.AddSync(os.Stdout)
	// 创建多个 WriteSyncer
	writeSyncer := zapcore.NewMultiWriteSyncer(fileSyncer, consoleSyncer)
	// 4.6 通过GetLevelEnabler获取自定义的LevelEnabler
	levelEnabler := GetLevelEnabler()
	// 4.7 通过Encoder、WriteSyncer、LevelEnabler创建一个core
	newCore := zapcore.NewCore(encoder, writeSyncer, levelEnabler)
	// 5 传递 newCore New一个logger
	//  zap.AddCaller(): 输出文件名和行号
	//  zap.Fields: 假如每条日志中需要携带公用的信息，可以在这里进行添加
	Logger = zap.New(newCore)
}

// getEncoder 自定义的Encoder    4.1
//    打开zapcore的源码，见图“zapcore-Encoder”：发现其中有两个new Encoder的func：
//          NewConsoleEncoder(console_encoder.go)
//          NewJSONEncoder(json_encoder.go)
//    这两个func都需要传递一个EncoderConfig的变量，而zap中已经给我们提供了几种获取EncoderConfig的方式
//          zap.NewProductionEncoderConfig()
//          zap.NewDevelopmentEncoderConfig()
//    在这里我直接把zap.NewProductionEncoderConfig()源码中的部分黏贴过来
func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(
		zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,      // 默认换行符"\n"
			EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 日志等级序列为小写字符串，如:InfoLevel被序列化为 "info"
			EncodeTime:     zapcore.EpochTimeEncoder,       // 日志时间格式显示
			EncodeDuration: zapcore.SecondsDurationEncoder, // 时间序列化，Duration为经过的浮点秒数
			EncodeCaller:   zapcore.ShortCallerEncoder,     // 日志行号显示
		})
}

// GetWriteSyncer 自定义的WriteSyncer 4.3
func getFileSyncer() zapcore.WriteSyncer {
	logsPath := "logs/agent.log"
	dir := filepath.Dir(logsPath)
	// 创建目录（包括父级目录）
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	file, err := os.Create(logsPath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	return zapcore.AddSync(file)
}

// GetLevelEnabler 自定义的LevelEnabler 4.5
func GetLevelEnabler() zapcore.Level {
	return zapcore.InfoLevel // 只会打印出info及其以上级别的日志
}
