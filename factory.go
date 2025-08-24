package embedder

import (
	"fmt"
	"sync"
)

// Factory 嵌入服务工厂
type Factory struct {
	providers map[string]ProviderFunc
	mu        sync.RWMutex
	logger    *Logger
}

// NewFactory 创建新的工厂实例
func NewFactory() *Factory {
	factory := &Factory{
		providers: make(map[string]ProviderFunc),
		logger:    NewLogger("embedder-factory"),
	}
	
	// 注册默认的 Ollama provider
	factory.RegisterProvider("ollama", func(config Config) (Embedder, error) {
		return NewOllamaEmbedder(config)
	})
	
	return factory
}

// Create 根据provider名称创建嵌入服务
func (f *Factory) Create(provider string) (Embedder, error) {
	f.mu.RLock()
	providerFunc, exists := f.providers[provider]
	f.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	
	// 使用默认配置，但设置正确的provider
	config := DefaultConfig
	config.Provider = provider
	
	f.logger.Info("创建嵌入服务", String("provider", provider))
	return providerFunc(config)
}

// CreateWithConfig 使用指定配置创建嵌入服务
func (f *Factory) CreateWithConfig(config Config) (Embedder, error) {
	f.mu.RLock()
	providerFunc, exists := f.providers[config.Provider]
	f.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
	
	f.logger.Info("创建嵌入服务", 
		String("provider", config.Provider),
		String("model", config.Model))
	return providerFunc(config)
}

// RegisterProvider 注册新的provider
func (f *Factory) RegisterProvider(name string, provider ProviderFunc) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if _, exists := f.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	
	f.providers[name] = provider
	f.logger.Info("注册新provider", String("name", name))
	return nil
}

// ListProviders 列出所有可用的provider
func (f *Factory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	providers := make([]string, 0, len(f.providers))
	for name := range f.providers {
		providers = append(providers, name)
	}
	return providers
}

// 全局工厂实例
var defaultFactory = NewFactory()

// New 创建嵌入服务的便捷方法，使用链式调用配置
func New(provider string) *EmbedderBuilder {
	return &EmbedderBuilder{
		config: EmbedderConfig{
			config: Config{
				Provider: provider,
				BaseURL:  DefaultConfig.BaseURL,
				Model:    DefaultConfig.Model,
				Timeout:  DefaultConfig.Timeout,
				Options:  make(map[string]interface{}),
			},
		},
		factory: defaultFactory,
	}
}

// EmbedderBuilder 嵌入服务构建器，支持链式调用
type EmbedderBuilder struct {
	config  EmbedderConfig
	factory *Factory
}

// WithBaseURL 设置基础URL
func (b *EmbedderBuilder) WithBaseURL(baseURL string) *EmbedderBuilder {
	b.config.WithBaseURL(baseURL)
	return b
}

// WithModel 设置模型名称
func (b *EmbedderBuilder) WithModel(model string) *EmbedderBuilder {
	b.config.WithModel(model)
	return b
}

// WithTimeout 设置超时时间
func (b *EmbedderBuilder) WithTimeout(timeout interface{}) *EmbedderBuilder {
	// 这里简化处理，实际使用中可以支持更多类型
	return b
}

// WithOption 设置自定义选项
func (b *EmbedderBuilder) WithOption(key string, value interface{}) *EmbedderBuilder {
	b.config.WithOption(key, value)
	return b
}

// LoadConfig 从YAML文件加载配置
func (b *EmbedderBuilder) LoadConfig(path string) error {
	return b.config.LoadConfig(path)
}

// Build 构建嵌入服务
func (b *EmbedderBuilder) Build() (Embedder, error) {
	return b.factory.CreateWithConfig(b.config.GetConfig())
}

// CreateEmbedder 直接创建嵌入服务的便捷方法
func CreateEmbedder(provider string) (Embedder, error) {
	return defaultFactory.Create(provider)
}

// CreateEmbedderWithConfig 使用配置创建嵌入服务的便捷方法
func CreateEmbedderWithConfig(config Config) (Embedder, error) {
	return defaultFactory.CreateWithConfig(config)
}

// RegisterProvider 注册provider的便捷方法
func RegisterProvider(name string, provider ProviderFunc) error {
	return defaultFactory.RegisterProvider(name, provider)
}

// ListProviders 列出providers的便捷方法
func ListProviders() []string {
	return defaultFactory.ListProviders()
}