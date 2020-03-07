package main

import (
	"strings"
	"sync"

	. "github.com/xaxys/oasis/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var timeFormat string
var loggerLock sync.Mutex
var loggerCore zapcore.Core
var serverLogger *zap.SugaredLogger

func getLogger() Logger {
	if loggerCore == nil {
		loggerLock.Lock()
		if loggerCore == nil {
			loggerCore = newLoggerCore(ServerConfig.GetString("LogPath"), ServerConfig.GetString("LogLevel"))
		}
		loggerLock.Unlock()
	}
	if serverLogger == nil {
		loggerLock.Lock()
		if serverLogger == nil {
			if ServerConfig.GetBool("DebugMode") {
				caller := zap.AddCaller()
				development := zap.Development()
				serverLogger = zap.New(loggerCore, caller, development).Sugar()
			} else {
				serverLogger = zap.New(loggerCore).Sugar()
			}
			serverLogger.Info("ServerLogger initialized successfully")
		}
		loggerLock.Unlock()
	}

	return serverLogger
}

type oasisLogger struct {
	*zap.SugaredLogger
}

func GetPluginLogger(name string) Logger {
	var pluginLogger *zap.SugaredLogger
	field := zap.Fields(zap.String("plugin", name))
	if ServerConfig.GetBool("DebugMode") {
		caller := zap.AddCaller()
		development := zap.Development()
		pluginLogger = zap.New(loggerCore, caller, development, field).Sugar()
	} else {
		pluginLogger = zap.New(loggerCore, field).Sugar()
	}

	return &oasisLogger{
		pluginLogger,
	}
}

// logpath 日志文件路径
// loglevel 日志级别
func newLoggerCore(logpath string, loglevel string) zapcore.Core {

	hook := lumberjack.Logger{
		Filename:   logpath, // 日志文件路径
		MaxSize:    2,       // megabytes
		MaxBackups: 300,     // 最多保留300个备份
		MaxAge:     365,     // days
		Compress:   true,    // 是否压缩 disabled by default
	}

	jsonEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 大写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	var level zapcore.Level
	switch strings.ToLower(loglevel) {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(jsonEncoderConfig),                                                 // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook), zapcore.AddSync(getConsolePrinter())), // 打印到文件
		level, // 日志级别
	)

	return core
}
