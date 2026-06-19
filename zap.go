package zap

import (
	"context"
	"os"

	"github.com/xudefa/go-boot/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapOption 定义 zap 日志适配器的配置选项，采用函数式选项模式。
// 通过传入一系列 ZapOption 来配置 NewZapAdapter 的行为，
// 支持链式调用和按需定制。
type ZapOption func(*ZapAdapter)

// WithZapLevel 设置日志级别，控制哪些级别的日志会被输出。
// 可用的级别从低到高包括：DebugLevel、InfoLevel、WarnLevel、
// ErrorLevel、DPanicLevel、PanicLevel、FatalLevel。
// 默认级别为 InfoLevel，低于该级别的日志将被过滤。
func WithZapLevel(level log.Level) ZapOption {
	return func(a *ZapAdapter) {
		a.level = level
	}
}

// WithZapFormat 设置日志输出格式，支持 "json"、"console" 和 "text" 三种格式。
// - "json": JSON 格式输出（默认），适合生产环境日志收集
// - "console" 或 "text": 控制台友好的彩色输出，适合本地开发调试
func WithZapFormat(format string) ZapOption {
	return func(a *ZapAdapter) {
		a.format = format
	}
}

// WithZapTimeFormat 设置日志时间戳的格式，使用 Go 标准时间格式布局。
// 默认格式为 "2006-01-02 15:04:05"，即年月日 时分秒。
// 可通过此选项自定义为其他格式，如 "2006-01-02" 仅显示日期。
func WithZapTimeFormat(timeFormat string) ZapOption {
	return func(a *ZapAdapter) {
		a.timeFormat = timeFormat
	}
}

// WithZapAddCaller 设置是否在日志中记录调用位置（文件名和行号）。
// 开启后每条日志会附加上调用 log 方法的源代码位置，
// 便于生产环境问题排查，但会带来轻微性能开销。
func WithZapAddCaller(addCaller bool) ZapOption {
	return func(a *ZapAdapter) {
		a.addCaller = addCaller
	}
}

// WithZapCallerSkip 设置调用者跳过的栈帧层数，默认跳过 1 层。
// 当存在多层包装时（如适配器在内部又调用了其他函数），
// 需要通过增加跳过帧数来定位到真正的业务调用处。
func WithZapCallerSkip(callerSkip int) ZapOption {
	return func(a *ZapAdapter) {
		a.callerSkip = callerSkip
	}
}

// WithZapOutput 设置日志输出的 WriteSyncer，支持同步写入操作。
// 可以传入标准输出、文件、网络连接等实现了 WriteSyncer 接口的对象。
// 注意此为灵活的底层设置，更快捷的文件输出可使用 WithZapOutputPath。
func WithZapOutput(output zapcore.WriteSyncer) ZapOption {
	return func(a *ZapAdapter) {
		a.output = output
	}
}

// WithZapOutputPath 设置日志文件输出路径，将日志写入指定文件。
// 如果文件不存在则创建，以追加模式写入，权限为 0644。
// 与 WithZapOutput 不同，此选项直接根据路径创建文件输出目标。
func WithZapOutputPath(path string) ZapOption {
	return func(a *ZapAdapter) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		a.output = zapcore.AddSync(f)
	}
}

// WithZapEncoder 设置自定义日志编码器，覆盖默认的 JSON 或 Console 编码器。
// 当默认的编码格式（JSON/Console）无法满足需求时，
// 可通过此选项传入自定义的 zapcore.Encoder 实现完全控制日志输出格式。
func WithZapEncoder(encoder zapcore.Encoder) ZapOption {
	return func(a *ZapAdapter) {
		a.encoder = encoder
	}
}

// ZapAdapter 是 zap 日志适配器，实现 log.Logger 接口。
// 它将 go.uber.org/zap 强大的高性能日志库适配到 go-boot 框架的
// log.Logger 抽象接口，使得 zap 可以作为统一的日志实现被框架使用。
// 支持 JSON/Console 输出格式、级别控制、调用者追踪等功能。
type ZapAdapter struct {
	logger     *zap.SugaredLogger
	level      log.Level
	format     string
	timeFormat string
	addCaller  bool
	callerSkip int
	encoder    zapcore.Encoder
	output     zapcore.WriteSyncer
}

// NewZapAdapter 创建 zap 日志适配器实例，通过函数式选项进行配置。
// 创建流程：
//  1. 使用默认配置初始化适配器（InfoLevel、JSON 格式、标准输出）
//  2. 依次应用传入的选项函数修改配置
//  3. 根据配置创建编码器和日志核心
//  4. 组装 zap.Logger 并通过 Sugar() 获得便捷的 SugaredLogger
func NewZapAdapter(opts ...ZapOption) *ZapAdapter {
	a := &ZapAdapter{
		level:      log.InfoLevel,
		format:     "json",
		timeFormat: "2006-01-02 15:04:05",
		addCaller:  false,
		callerSkip: 1,
		output:     zapcore.AddSync(os.Stdout),
	}

	for _, opt := range opts {
		opt(a)
	}

	var encoder zapcore.Encoder
	encodeConfig := &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(a.timeFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if a.format == "text" || a.format == "console" {
		encoder = zapcore.NewConsoleEncoder(*encodeConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(*encodeConfig)
	}
	if a.encoder != nil {
		encoder = a.encoder
	}

	level := a.toZapLevel(a.level)
	core := zapcore.NewCore(
		encoder,
		a.output,
		level,
	)
	var zapOpts []zap.Option
	if a.addCaller {
		zapOpts = append(zapOpts, zap.AddCaller())
		if a.callerSkip > 0 {
			zapOpts = append(zapOpts, zap.AddCallerSkip(a.callerSkip))
		}
	}
	zapOpts = append(zapOpts, zap.Development())
	l := zap.New(core, zapOpts...)
	a.logger = l.Sugar()

	return a
}

// toZapLevel 将 go-boot 框架统一的日志级别 log.Level 转换为
// zap 框架内部的 zapcore.Level 类型，实现两个日志体系的桥接。
// 当遇到未知级别时，默认返回 InfoLevel 以保证框架稳定运行。
func (a *ZapAdapter) toZapLevel(level log.Level) zapcore.Level {
	switch level {
	case log.DebugLevel:
		return zapcore.DebugLevel
	case log.InfoLevel:
		return zapcore.InfoLevel
	case log.WarnLevel:
		return zapcore.WarnLevel
	case log.ErrorLevel:
		return zapcore.ErrorLevel
	case log.DPanicLevel:
		return zapcore.DPanicLevel
	case log.PanicLevel:
		return zapcore.PanicLevel
	case log.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// toZapFields 将 go-boot 的日志键值对列表 []log.KeyValue 转换为
// zap.Field 的变长参数列表（以 []any 形式），以便传递给
// SugaredLogger 的各类日志方法。
func (a *ZapAdapter) toZapFields(keys []log.KeyValue) []any {
	var fields []any
	for _, kv := range keys {
		fields = append(fields, zap.Any(kv.Key, kv.Value))
	}
	return fields
}

// log 是内部统一的日志记录方法，根据传入的日志级别将消息和键值对
// 分发到 zap.SugaredLogger 对应的级别方法上。
// ctx 参数预留用于后续支持日志链路追踪上下文传递。
func (a *ZapAdapter) log(ctx context.Context, level log.Level, msg string, keys []log.KeyValue) {
	fields := a.toZapFields(keys)

	switch level {
	case log.DebugLevel:
		a.logger.Debugw(msg, fields...)
	case log.InfoLevel:
		a.logger.Infow(msg, fields...)
	case log.WarnLevel:
		a.logger.Warnw(msg, fields...)
	case log.ErrorLevel:
		a.logger.Errorw(msg, fields...)
	case log.DPanicLevel:
		a.logger.DPanicw(msg, fields...)
	case log.PanicLevel:
		a.logger.Panicw(msg, fields...)
	case log.FatalLevel:
		a.logger.Fatalw(msg, fields...)
	}
}

// Debug 记录调试级别的日志，用于开发和排查问题时的详细信息输出。
// 在生产环境中通常会被过滤掉以减少日志量。
func (a *ZapAdapter) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DebugLevel, msg, keys)
}

// Info 记录信息级别的日志，用于正常的业务运行信息记录，
// 如请求处理完成、服务启动等常规事件。
func (a *ZapAdapter) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.InfoLevel, msg, keys)
}

// Warn 记录警告级别的日志，当系统出现非关键性异常时使用。
// 表示可能存在问题但服务仍能正常运行，需要关注但不需立即处理。
func (a *ZapAdapter) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.WarnLevel, msg, keys)
}

// Error 记录错误级别的日志，用于系统出现需要关注的异常情况。
// 表示某个操作执行失败，但不影响整个应用的继续运行。
func (a *ZapAdapter) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.ErrorLevel, msg, keys)
}

// DPanic 记录致命级别的日志并触发 panic，但仅在开发环境中生效。
// 在开发模式下遇到严重错误时立即中断程序执行，
// 而在生产环境中则降级为 Error 级别的日志记录。
func (a *ZapAdapter) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DPanicLevel, msg, keys)
}

// Panic 记录日志后调用 panic 中止当前控制流程。
// 适用于遇到了程序无法恢复的严重错误场景。
func (a *ZapAdapter) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.PanicLevel, msg, keys)
}

// Fatal 记录日志后调用 os.Exit(1) 立即终止程序运行。
// 用于最严重的系统级错误场景，如无法连接关键依赖服务。
func (a *ZapAdapter) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.FatalLevel, msg, keys)
}

// Sync 刷新日志缓冲区，将 buffer 中的日志数据强制刷写到输出目标。
// 在应用优雅关闭时应调用此方法确保所有日志都已持久化，
// 避免因进程退出导致日志丢失。
func (a *ZapAdapter) Sync() error {
	return a.logger.Sync()
}

// With 返回一个新的日志记录器，该记录器会在每条日志中自动附加指定的键值对。
// 适用于需要在上下文范围内固定携带某些字段（如请求 ID、用户 ID 等）的场景。
// 返回的新适配器共享原有适配器的其他配置（级别、格式等），
// 但拥有独立扩展的字段集，互不影响。
func (a *ZapAdapter) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
	fields := a.toZapFields(keys)
	return &ZapAdapter{
		logger:     a.logger.With(fields...),
		level:      a.level,
		format:     a.format,
		timeFormat: a.timeFormat,
		addCaller:  a.addCaller,
		callerSkip: a.callerSkip,
		encoder:    a.encoder,
		output:     a.output,
	}
}
