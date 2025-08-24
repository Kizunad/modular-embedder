# @embedder/ 嵌入服务模块

通用嵌入服务模块，支持多种提供者，提供统一的API接口。

## 特性

- 🔌 多提供者支持 (当前支持 Ollama)
- ⚙️ 灵活配置 (WithXXX 方法 + YAML 文件)
- 🧪 完整测试覆盖
- 📝 简化日志输出
- 🚀 链式调用API

## 快速开始

### 链式调用方式

```go
import "embedder"

// 使用默认配置
embedder := embedder.New("ollama").Build()

// 自定义配置
embedder := embedder.New("ollama").
    WithBaseURL("http://localhost:11434").
    WithModel("qwen2.5:7b").
    Build()
```

### YAML 配置方式

```yaml
# embedder.yaml
provider: "ollama"
base_url: "http://localhost:11434"
model: "qwen2.5:7b"
timeout: "30s"
options:
  temperature: 0.7
```

```go
builder := embedder.New("ollama")
err := builder.LoadConfig("embedder.yaml")
embedder := builder.Build()
```

## API 使用

```go
ctx := context.Background()

// 单个文本嵌入
embedding, err := embedder.EmbedSingle(ctx, "Hello, world!")

// 批量文本嵌入
texts := []string{"text1", "text2", "text3"}
embeddings, err := embedder.Embed(ctx, texts)

// 分批处理大量文本
embeddings, err := embedder.BatchEmbed(ctx, texts, 10) // 每批10个

// 获取模型信息
model := embedder.GetModel()
dimension := embedder.GetDimension()

// 健康检查
err := embedder.Health(ctx)
```

## 扩展新的提供者

```go
func NewCustomEmbedder(config embedder.Config) (embedder.Embedder, error) {
    // 实现自定义嵌入服务
    return &CustomEmbedder{}, nil
}

// 注册自定义提供者
embedder.RegisterProvider("custom", NewCustomEmbedder)

// 使用自定义提供者
embedder := embedder.New("custom").Build()
```