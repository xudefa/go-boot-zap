// Package zap 包含 go.uber.org/zap 日志库的适配器单元测试。
// 测试覆盖适配器的创建、选项配置、不同日志级别的调用、
// 字段扩展、缓冲区同步以及接口级别转换等功能。
package zap

import (
	"context"
	"testing"

	"github.com/xudefa/go-boot/log"
)

// TestNewZapAdapter 验证使用基本选项创建 ZapAdapter 是否成功，
// 确保在指定级别、格式和时间格式的情况下适配器能正常初始化。
func TestNewZapAdapter(t *testing.T) {
	adapter := NewZapAdapter(
		WithZapLevel(log.DebugLevel),
		WithZapFormat("json"),
		WithZapTimeFormat("2006-01-02"),
	)
	if adapter == nil {
		t.Error("NewZapAdapter() returned nil")
	}
}

// TestNewZapAdapterWithOptions 通过表驱动测试验证各个选项函数（WithZapLevel、
// WithZapAddCaller、WithZapCallerSkip、WithZapTimeFormat）单独使用时
// 都能正确创建适配器实例，确保每个选项的独立可用性。
func TestNewZapAdapterWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  ZapOption
	}{
		{"WithZapLevel", WithZapLevel(log.DebugLevel)},

		{"WithZapAddCaller", WithZapAddCaller(true)},
		{"WithZapCallerSkip", WithZapCallerSkip(2)},
		{"WithZapTimeFormat", WithZapTimeFormat("2006-01-02")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewZapAdapter(tt.opt)
			if adapter == nil {
				t.Error("option failed")
			}
		})
	}
}

// TestZapAdapterLogLevels 验证适配器在默认配置下能正确调用各个级别的日志方法。
// 依次调用 Debug、Info、Warn、Error 四个级别，确保不 panic 且调用链路正常。
// 注意：此测试不验证输出内容，仅验证调用过程无异常。
func TestZapAdapterLogLevels(t *testing.T) {
	adapter := NewZapAdapter()
	ctx := context.Background()

	adapter.Debug(ctx, "debug message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Info(ctx, "info message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Warn(ctx, "warn message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Error(ctx, "error message", log.KeyValue{Key: "k", Value: "v"})
}

// TestZapAdapterWith 验证 With 方法返回的新适配器实例不为 nil，
// 确保通过 With 扩展上下文字段的功能正常工作。
func TestZapAdapterWith(t *testing.T) {
	adapter := NewZapAdapter()
	ctx := context.Background()

	newAdapter := adapter.With(ctx, log.KeyValue{Key: "k", Value: "v"})
	if newAdapter == nil {
		t.Error("With() returned nil")
	}
}

// TestZapAdapterSync 验证 Sync 方法能正常调用且不 panic。
// Sync 用于强制刷新日志缓冲区，确保日志被写入输出目标。
func TestZapAdapterSync(t *testing.T) {
	adapter := NewZapAdapter(WithZapFormat("json"))
	_ = adapter.Sync()
}

// TestZapAdapterImplementsInterface 在编译时验证 ZapAdapter 类型
// 是否实现了 log.Logger 接口，利用 Go 的类型系统进行接口兼容性检查。
func TestZapAdapterImplementsInterface(t *testing.T) {
	var _ log.Logger = (*ZapAdapter)(nil)
}

// TestToZapLevel 验证 toZapLevel 方法能正确地将 go-boot 的 log.Level
// 映射为 zap 的 zapcore.Level。覆盖所有日志级别（Debug、Info、Warn、
// Error、DPanic、Panic、Fatal），确保映射关系正确。
func TestToZapLevel(t *testing.T) {
	adapter := &ZapAdapter{}

	tests := []struct {
		input    log.Level
		expected string
	}{
		{log.DebugLevel, "debug"},
		{log.InfoLevel, "info"},
		{log.WarnLevel, "warn"},
		{log.ErrorLevel, "error"},
		{log.DPanicLevel, "dpanic"},
		{log.PanicLevel, "panic"},
		{log.FatalLevel, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			lvl := adapter.toZapLevel(tt.input)
			if lvl.String() != tt.expected {
				t.Errorf("toZapLevel() = %v, want %v", lvl, tt.expected)
			}
		})
	}
}

// TestToZapFields 验证 toZapFields 方法能正确地将 log.KeyValue 键值对列表
// 转换为普通的字段列表，测试包含字符串、整数和布尔三种类型的值。
func TestToZapFields(t *testing.T) {
	adapter := &ZapAdapter{}
	keys := []log.KeyValue{
		{Key: "k1", Value: "v1"},
		{Key: "k2", Value: 123},
		{Key: "k3", Value: true},
	}

	fields := adapter.toZapFields(keys)
	if len(fields) != 3 {
		t.Errorf("toZapFields() returned %d fields, want 3", len(fields))
	}
}
