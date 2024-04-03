// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package log is a log package used by onex project.
//
//nolint:interfacebloat
package log

import (
	"sync"
	"time"

	krtlog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

type Field = zapcore.Field

// Logger 定义了 onex 项目的日志接口. 该接口只包含了支持的日志记录方法.
type Logger interface {
	Debugf(format string, args ...any)
	Debugw(msg string, keyvals ...any)
	Infof(format string, args ...any)
	Infow(msg string, keyvals ...any)
	Warnf(format string, args ...any)
	Warnw(msg string, keyvals ...any)
	Errorf(format string, args ...any)
	Errorw(err error, msg string, keyvals ...any)
	Panicf(format string, args ...any)
	Panicw(msg string, keyvals ...any)
	Fatalf(format string, args ...any)
	Fatalw(msg string, keyvals ...any)
	With(fields ...Field) Logger
	AddCallerSkip(skip int) Logger
	Sync()

	// integrate other loggers
	krtlog.Logger
	gormlogger.Interface
}

// zapLogger 是 Logger 接口的具体实现. 它底层封装了 zap.Logger.
type zapLogger struct {
	z    *zap.Logger
	opts *Options
}

// 确保 zapLogger 实现了 Logger 接口. 以下变量赋值，可以使错误在编译期被发现.
var _ Logger = (*zapLogger)(nil)

var (
	mu sync.Mutex

	// std 定义了默认的全局 Logger.
	std = NewLogger(NewOptions())
)

// Init 使用指定的选项初始化 Logger.
func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()

	std = NewLogger(opts)
}

// NewLogger 根据传入的 opts 创建 Logger.
func NewLogger(opts *Options) *zapLogger {
	if opts == nil {
		opts = NewOptions()
	}

	// 将文本格式的日志级别，例如 info 转换为 zapcore.Level 类型以供后面使用
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		// 如果指定了非法的日志级别，则默认使用 info 级别
		zapLevel = zapcore.InfoLevel
	}

	// 创建一个默认的 encoder 配置
	encoderConfig := zap.NewProductionEncoderConfig()
	// 自定义 MessageKey 为 message，message 语义更明确
	encoderConfig.MessageKey = "message"
	// 自定义 TimeKey 为 timestamp，timestamp 语义更明确
	encoderConfig.TimeKey = "timestamp"
	// 指定时间序列化函数，将时间序列化为 `2006-01-02 15:04:05.000` 格式，更易读
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	// 指定 time.Duration 序列化函数，将 time.Duration 序列化为经过的毫秒数的浮点数
	// 毫秒数比默认的秒数更精确
	encoderConfig.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendFloat64(float64(d) / float64(time.Millisecond))
	}
	// when output to local path, with color is forbidden
	if opts.Format == "console" && opts.EnableColor {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 创建构建 zap.Logger 需要的配置
	cfg := &zap.Config{
		// 是否在日志中显示调用日志所在的文件和行号，例如：`"caller":"onex/onex.go:75"`
		DisableCaller: opts.DisableCaller,
		// 是否禁止在 panic 及以上级别打印堆栈信息
		DisableStacktrace: opts.DisableStacktrace,
		// 指定日志级别
		Level: zap.NewAtomicLevelAt(zapLevel),
		// 指定日志显示格式，可选值：console, json
		Encoding:      opts.Format,
		EncoderConfig: encoderConfig,
		// 指定日志输出位置
		OutputPaths: opts.OutputPaths,
		// 设置 zap 内部错误输出位置
		ErrorOutputPaths: []string{"stderr"},
	}

	// 使用 cfg 创建 *zap.Logger 对象
	z, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(2))
	if err != nil {
		panic(err)
	}
	logger := &zapLogger{z: z, opts: opts}

	// 把标准库的 log.Logger 的 info 级别的输出重定向到 zap.Logger
	zap.RedirectStdLog(z)

	return logger
}

func Default() Logger {
	return std
}

func (l *zapLogger) Options() *Options {
	return l.opts
}

// Sync 调用底层 zap.Logger 的 Sync 方法，将缓存中的日志刷新到磁盘文件中. 主程序需要在退出前调用 Sync.
func Sync() { std.Sync() }

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}

// Debugf 输出 debug 级别的日志.
func Debugf(format string, args ...any) {
	std.Debugf(format, args...)
}

func (l *zapLogger) Debugf(format string, args ...any) {
	l.z.Sugar().Debugf(format, args...)
}

// Debugw 输出 debug 级别的日志.
func Debugw(msg string, keyvals ...any) {
	std.Debugw(msg, keyvals...)
}

func (l *zapLogger) Debugw(msg string, keyvals ...any) {
	l.z.Sugar().Debugw(msg, keyvals...)
}

// Infof 输出 info 级别的日志.
func Infof(format string, args ...any) {
	std.Infof(format, args...)
}

func (l *zapLogger) Infof(msg string, keyvals ...any) {
	l.z.Sugar().Infof(msg, keyvals...)
}

// Infow 输出 info 级别的日志.
func Infow(msg string, keyvals ...any) {
	std.Infow(msg, keyvals...)
}

func (l *zapLogger) Infow(msg string, keyvals ...any) {
	l.z.Sugar().Infow(msg, keyvals...)
}

// Warnf 输出 warning 级别的日志.
func Warnf(format string, args ...any) {
	std.Warnf(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...any) {
	l.z.Sugar().Warnf(format, args...)
}

// Warnw 输出 warning 级别的日志.
func Warnw(msg string, keyvals ...any) {
	std.Warnw(msg, keyvals...)
}

func (l *zapLogger) Warnw(msg string, keyvals ...any) {
	l.z.Sugar().Warnw(msg, keyvals...)
}

// Errorf 输出 error 级别的日志.
func Errorf(format string, args ...any) {
	std.Errorf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...any) {
	l.z.Sugar().Errorf(format, args...)
}

// Errorw 输出 error 级别的日志.
func Errorw(err error, msg string, keyvals ...any) {
	std.Errorw(err, msg, keyvals...)
}

func (l *zapLogger) Errorw(err error, msg string, keyvals ...any) {
	l.z.Sugar().Errorw(msg, append(keyvals, "err", err)...)
}

// Panicf 输出 panic 级别的日志.
func Panicf(format string, args ...any) {
	std.Panicf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...any) {
	l.z.Sugar().Panicf(format, args...)
}

// Panicw 输出 panic 级别的日志.
func Panicw(msg string, keyvals ...any) {
	std.Panicw(msg, keyvals...)
}

func (l *zapLogger) Panicw(msg string, keyvals ...any) {
	l.z.Sugar().Panicw(msg, keyvals...)
}

// Fatalf 输出 fatal 级别的日志.
func Fatalf(format string, args ...any) {
	std.Fatalf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...any) {
	l.z.Sugar().Fatalf(format, args...)
}

// Fatalw 输出 fatal 级别的日志.
func Fatalw(msg string, keyvals ...any) {
	std.Fatalw(msg, keyvals...)
}

func (l *zapLogger) Fatalw(msg string, keyvals ...any) {
	l.z.Sugar().Fatalw(msg, keyvals...)
}

func With(fields ...Field) Logger {
	return std.With(fields...)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (l *zapLogger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return l
	}

	lc := l.clone()
	lc.z = lc.z.With(fields...)
	return lc
}

func AddCallerSkip(skip int) Logger {
	return std.AddCallerSkip(skip)
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents zap from always
// reporting the wrapper code as the caller.
func (l *zapLogger) AddCallerSkip(skip int) Logger {
	lc := l.clone()
	lc.z = lc.z.WithOptions(zap.AddCallerSkip(skip))
	return lc
}

// clone 深度拷贝 zapLogger.
func (l *zapLogger) clone() *zapLogger {
	copied := *l
	return &copied
}
