package embedder

import (
	"context"
	"time"
)

// Embedder 嵌入服务核心接口
// 提供文本嵌入功能，支持单个、批量和分批处理
type Embedder interface {
	// Embed 批量嵌入多个文本
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	
	// EmbedSingle 嵌入单个文本
	EmbedSingle(ctx context.Context, text string) ([]float32, error)
	
	// BatchEmbed 分批处理大量文本，避免单次请求过大
	BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error)
	
	// GetDimension 获取嵌入向量维度
	GetDimension() int
	
	// GetModel 获取当前使用的模型名称
	GetModel() string
	
	// Health 健康检查，确认服务可用
	Health(ctx context.Context) error
}

// EmbedderFactory 嵌入服务工厂接口
// 支持通过provider名称创建不同的嵌入服务
type EmbedderFactory interface {
	// Create 根据provider名称创建嵌入服务
	Create(provider string) (Embedder, error)
	
	// RegisterProvider 注册新的provider
	RegisterProvider(name string, provider ProviderFunc) error
	
	// ListProviders 列出所有可用的provider
	ListProviders() []string
}

// ProviderFunc provider创建函数类型
type ProviderFunc func(config Config) (Embedder, error)

// Config 嵌入服务配置
type Config struct {
	Provider string                 `yaml:"provider"`
	BaseURL  string                 `yaml:"base_url"`
	Model    string                 `yaml:"model"`
	Timeout  time.Duration          `yaml:"timeout"`
	Options  map[string]interface{} `yaml:"options"`
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Provider: "ollama",
	BaseURL:  "http://localhost:11434",
	Model:    "qwen2.5:7b",
	Timeout:  30 * time.Second,
	Options:  make(map[string]interface{}),
}