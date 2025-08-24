package main

import (
	"context"
	"fmt"
	"log"

	embedder "embedder"
)

func main() {
	// 示例1: 链式调用方式
	fmt.Println("=== 示例1: 链式调用方式 ===")
	
	e1, err := embedder.New("ollama").
		WithBaseURL("http://localhost:11434").
		WithModel("qwen2.5:7b").
		Build()
	
	if err != nil {
		log.Printf("创建embedder失败: %v", err)
		return
	}

	fmt.Printf("模型: %s, 维度: %d\n", e1.GetModel(), e1.GetDimension())

	// 测试单个文本嵌入
	ctx := context.Background()
	embedding, err := e1.EmbedSingle(ctx, "Hello, world!")
	if err != nil {
		log.Printf("嵌入失败: %v", err)
	} else {
		fmt.Printf("单个文本嵌入结果 (前3个值): [%.3f, %.3f, %.3f...]\n", 
			embedding[0], embedding[1], embedding[2])
	}

	// 测试批量嵌入
	texts := []string{"你好", "世界", "人工智能"}
	embeddings, err := e1.Embed(ctx, texts)
	if err != nil {
		log.Printf("批量嵌入失败: %v", err)
	} else {
		fmt.Printf("批量嵌入 %d 个文本成功\n", len(embeddings))
	}

	fmt.Println("\n=== 示例2: YAML配置方式 ===")

	// 创建配置文件
	configContent := `provider: "ollama"
base_url: "http://localhost:11434"
model: "qwen2.5:7b"
timeout: "30s"
options:
  temperature: 0.7`

	// 这里简化演示，实际使用中应该从文件加载
	fmt.Println("配置内容:")
	fmt.Println(configContent)

	fmt.Println("\n=== 示例3: 提供者管理 ===")

	// 列出所有可用的提供者
	providers := embedder.ListProviders()
	fmt.Printf("可用的提供者: %v\n", providers)

	// 注册自定义提供者
	err = embedder.RegisterProvider("mock", func(config embedder.Config) (embedder.Embedder, error) {
		return &MockEmbedder{model: config.Model}, nil
	})
	if err != nil {
		log.Printf("注册提供者失败: %v", err)
	} else {
		fmt.Println("成功注册 mock 提供者")
	}

	// 使用自定义提供者
	mockEmbedder, err := embedder.CreateEmbedder("mock")
	if err != nil {
		log.Printf("创建mock embedder失败: %v", err)
	} else {
		fmt.Printf("Mock embedder - 模型: %s, 维度: %d\n", 
			mockEmbedder.GetModel(), mockEmbedder.GetDimension())
		
		mockEmbedding, _ := mockEmbedder.EmbedSingle(ctx, "test")
		fmt.Printf("Mock嵌入结果: %v\n", mockEmbedding)
	}

	fmt.Println("\n=== 示例4: 分批处理 ===")

	// 创建大量文本用于演示分批处理
	largeTexts := make([]string, 20)
	for i := range largeTexts {
		largeTexts[i] = fmt.Sprintf("文本 %d", i+1)
	}

	if mockEmbedder != nil {
		batchEmbeddings, err := mockEmbedder.BatchEmbed(ctx, largeTexts, 5)
		if err != nil {
			log.Printf("分批嵌入失败: %v", err)
		} else {
			fmt.Printf("分批处理 %d 个文本，每批 5 个，共得到 %d 个嵌入向量\n", 
				len(largeTexts), len(batchEmbeddings))
		}
	}
}

// MockEmbedder 用于演示的模拟嵌入服务
type MockEmbedder struct {
	model string
}

func (m *MockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range result {
		result[i] = []float32{0.1 + float32(i)*0.01, 0.2 + float32(i)*0.01, 0.3 + float32(i)*0.01}
	}
	return result, nil
}

func (m *MockEmbedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	// 简单的模拟：根据文本长度生成不同的向量
	length := float32(len(text)) * 0.01
	return []float32{0.1 + length, 0.2 + length, 0.3 + length}, nil
}

func (m *MockEmbedder) BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
	return m.Embed(ctx, texts)
}

func (m *MockEmbedder) GetDimension() int {
	return 3
}

func (m *MockEmbedder) GetModel() string {
	if m.model == "" {
		return "mock-model"
	}
	return m.model
}

func (m *MockEmbedder) Health(ctx context.Context) error {
	return nil
}