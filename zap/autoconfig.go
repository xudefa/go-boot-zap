// Package zap 提供 Zap 日志适配器的自动配置。
//
// 当 zap.enabled=true 时自动启用，从 Environment 中读取 zap.level、zap.format、zap.time-format、
// zap.add-caller、zap.caller-skip、zap.output 等配置项，
// 创建并注册 Zap Logger Bean 到 IoC 容器中（Bean ID: zapLogger），实现 log.Logger 接口。
package zap

import (
	zapcore "github.com/xudefa/go-boot-zap"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/log"
)

// init 注册 Zap 自动配置，由 zap.enabled=true 条件控制。
func init() {
	boot.RegisterAutoConfig(&ZapAutoConfiguration{},
		condition.OnProperty("zap.enabled", "true"),
	)
}

// ZapAutoConfiguration Zap 日志适配器的自动配置。
//
// 从 Environment 中读取 zap.level、zap.format、zap.output 等配置项，
// 创建 Zap 日志适配器并注册到 IoC 容器中，实现 log.Logger 接口。
// 启用条件：zap.enabled=true
type ZapAutoConfiguration struct{}

// Configure 执行自动配置逻辑，创建 ZapLogger 并注册为 Bean。
func (z *ZapAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	opts := []zapcore.ZapOption{
		zapcore.WithZapLevel(log.ToLevel(env.GetString("zap.level", "info"))),
		zapcore.WithZapFormat(env.GetString("zap.format", "json")),
		zapcore.WithZapTimeFormat(env.GetString("zap.time-format", "2006-01-02 15:04:05")),
		zapcore.WithZapAddCaller(env.GetBool("zap.add-caller", false)),
		zapcore.WithZapCallerSkip(env.GetInt("zap.caller-skip", 1)),
	}
	if output := env.GetString("zap.output", ""); output != "" {
		opts = append(opts, zapcore.WithZapOutputPath(output))
	}

	logger := zapcore.NewZapAdapter(opts...)

	if err := ctx.Register("zapLogger",
		core.Bean(logger),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}

// 编译时检查 ZapAdapter 是否实现了 log.Logger 接口
var _ log.Logger = (*zapcore.ZapAdapter)(nil)
