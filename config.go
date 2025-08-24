package embedder

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// EmbedderConfig 嵌入服务配置管理器
type EmbedderConfig struct {
	config Config
}

// NewConfig 创建新的配置管理器
func NewConfig() *EmbedderConfig {
	return &EmbedderConfig{
		config: DefaultConfig,
	}
}

// WithProvider 设置提供者
func (c *EmbedderConfig) WithProvider(provider string) *EmbedderConfig {
	c.config.Provider = provider
	return c
}

// WithBaseURL 设置基础URL
func (c *EmbedderConfig) WithBaseURL(baseURL string) *EmbedderConfig {
	c.config.BaseURL = baseURL
	return c
}

// WithModel 设置模型名称
func (c *EmbedderConfig) WithModel(model string) *EmbedderConfig {
	c.config.Model = model
	return c
}

// WithTimeout 设置超时时间
func (c *EmbedderConfig) WithTimeout(timeout time.Duration) *EmbedderConfig {
	c.config.Timeout = timeout
	return c
}

// WithOption 设置自定义选项
func (c *EmbedderConfig) WithOption(key string, value interface{}) *EmbedderConfig {
	if c.config.Options == nil {
		c.config.Options = make(map[string]interface{})
	}
	c.config.Options[key] = value
	return c
}

// LoadConfig 从YAML文件加载配置
func (c *EmbedderConfig) LoadConfig(path string) error {
	config, err := LoadConfig(path)
	if err != nil {
		return err
	}
	c.config = *config
	return nil
}

// GetConfig 获取配置
func (c *EmbedderConfig) GetConfig() Config {
	return c.config
}

// LoadConfig 加载YAML配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.Provider == "" {
		config.Provider = DefaultConfig.Provider
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultConfig.BaseURL
	}
	if config.Model == "" {
		config.Model = DefaultConfig.Model
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig.Timeout
	}
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}

	return &config, nil
}