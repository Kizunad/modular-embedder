package embedder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if config == nil {
		t.Fatal("NewConfig returned nil")
	}
	
	cfg := config.GetConfig()
	if cfg.Provider != DefaultConfig.Provider {
		t.Errorf("Expected provider %s, got %s", DefaultConfig.Provider, cfg.Provider)
	}
}

func TestEmbedderConfigWithMethods(t *testing.T) {
	config := NewConfig().
		WithProvider("test").
		WithBaseURL("http://test.com").
		WithModel("test-model").
		WithTimeout(10 * time.Second).
		WithOption("key", "value")

	cfg := config.GetConfig()
	
	if cfg.Provider != "test" {
		t.Errorf("Expected provider 'test', got %s", cfg.Provider)
	}
	if cfg.BaseURL != "http://test.com" {
		t.Errorf("Expected baseURL 'http://test.com', got %s", cfg.BaseURL)
	}
	if cfg.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", cfg.Model)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", cfg.Timeout)
	}
	if cfg.Options["key"] != "value" {
		t.Errorf("Expected option key=value, got %v", cfg.Options["key"])
	}
}

func TestFactory(t *testing.T) {
	factory := NewFactory()
	
	// 测试列出providers
	providers := factory.ListProviders()
	found := false
	for _, p := range providers {
		if p == "ollama" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'ollama' provider to be registered by default")
	}
	
	// 测试注册新provider
	mockProvider := func(config Config) (Embedder, error) {
		return &MockEmbedder{}, nil
	}
	
	err := factory.RegisterProvider("mock", mockProvider)
	if err != nil {
		t.Errorf("Failed to register mock provider: %v", err)
	}
	
	// 测试重复注册
	err = factory.RegisterProvider("mock", mockProvider)
	if err == nil {
		t.Error("Expected error when registering duplicate provider")
	}
}

func TestNewBuilder(t *testing.T) {
	builder := New("test").
		WithBaseURL("http://test.com").
		WithModel("test-model").
		WithOption("key", "value")
	
	config := builder.config.GetConfig()
	if config.Provider != "test" {
		t.Errorf("Expected provider 'test', got %s", config.Provider)
	}
	if config.BaseURL != "http://test.com" {
		t.Errorf("Expected baseURL 'http://test.com', got %s", config.BaseURL)
	}
}

func TestOllamaEmbedderWithMockServer(t *testing.T) {
	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "0.1.0"}`))
		case "/api/embeddings":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"embedding": [0.1, 0.2, 0.3, 0.4]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 创建配置
	config := Config{
		Provider: "ollama",
		BaseURL:  server.URL,
		Model:    "test-model",
		Timeout:  5 * time.Second,
		Options:  make(map[string]interface{}),
	}

	// 创建embedder
	embedder, err := NewOllamaEmbedder(config)
	if err != nil {
		t.Fatalf("Failed to create OllamaEmbedder: %v", err)
	}

	// 测试基本信息
	if embedder.GetModel() != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", embedder.GetModel())
	}
	
	if embedder.GetDimension() != 4 {
		t.Errorf("Expected dimension 4, got %d", embedder.GetDimension())
	}

	// 测试健康检查
	ctx := context.Background()
	err = embedder.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 测试单个嵌入
	embedding, err := embedder.EmbedSingle(ctx, "test text")
	if err != nil {
		t.Errorf("EmbedSingle failed: %v", err)
	}
	if len(embedding) != 4 {
		t.Errorf("Expected embedding length 4, got %d", len(embedding))
	}

	// 测试批量嵌入
	texts := []string{"text1", "text2"}
	embeddings, err := embedder.Embed(ctx, texts)
	if err != nil {
		t.Errorf("Embed failed: %v", err)
	}
	if len(embeddings) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(embeddings))
	}

	// 测试分批嵌入
	batchEmbeddings, err := embedder.BatchEmbed(ctx, texts, 1)
	if err != nil {
		t.Errorf("BatchEmbed failed: %v", err)
	}
	if len(batchEmbeddings) != 2 {
		t.Errorf("Expected 2 batch embeddings, got %d", len(batchEmbeddings))
	}
}

func TestCreateEmbedder(t *testing.T) {
	// 测试不存在的provider
	_, err := CreateEmbedder("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent provider")
	}
}

func TestListProviders(t *testing.T) {
	providers := ListProviders()
	if len(providers) == 0 {
		t.Error("Expected at least one provider")
	}
}

// MockEmbedder 用于测试的模拟嵌入服务
type MockEmbedder struct{}

func (m *MockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range result {
		result[i] = []float32{0.1, 0.2, 0.3}
	}
	return result, nil
}

func (m *MockEmbedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *MockEmbedder) BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
	return m.Embed(ctx, texts)
}

func (m *MockEmbedder) GetDimension() int {
	return 3
}

func (m *MockEmbedder) GetModel() string {
	return "mock-model"
}

func (m *MockEmbedder) Health(ctx context.Context) error {
	return nil
}