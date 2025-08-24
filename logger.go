package embedder

import (
	"fmt"
	"log"
	"os"
)

// Logger 简化的日志接口，复制自zap的基础功能
type Logger struct {
	name string
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// NewLogger 创建新的日志记录器
func NewLogger(name string) *Logger {
	return &Logger{name: name}
}

// Named 创建命名的子日志记录器
func (l *Logger) Named(name string) *Logger {
	return &Logger{name: l.name + "." + name}
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields...)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields...)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, fields ...Field) {
	if os.Getenv("DEBUG") != "" {
		l.log("DEBUG", msg, fields...)
	}
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields...)
}

// log 内部日志记录方法
func (l *Logger) log(level, msg string, fields ...Field) {
	logMsg := fmt.Sprintf("[%s] %s: %s", level, l.name, msg)
	
	if len(fields) > 0 {
		logMsg += " {"
		for i, field := range fields {
			if i > 0 {
				logMsg += ", "
			}
			logMsg += fmt.Sprintf("%s: %v", field.Key, field.Value)
		}
		logMsg += "}"
	}
	
	log.Println(logMsg)
}