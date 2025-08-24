package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaEmbedder Ollama嵌入服务实现
type OllamaEmbedder struct {
	baseURL    string
	model      string
	httpClient *http.Client
	dimension  int
	logger     *Logger
}

// ollamaEmbedRequest Ollama嵌入请求格式
type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbedResponse Ollama嵌入响应格式
type ollamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// NewOllamaEmbedder 创建新的Ollama嵌入服务
func NewOllamaEmbedder(config Config) (*OllamaEmbedder, error) {
	logger := NewLogger("ollama-embedder")

	embedder := &OllamaEmbedder{
		baseURL: strings.TrimSuffix(config.BaseURL, "/"),
		model:   config.Model,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}

	// 测试连接并获取模型信息
	ctx := context.Background()
	if err := embedder.Health(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	// 获取嵌入维度
	if err := embedder.detectDimension(ctx); err != nil {
		return nil, fmt.Errorf("failed to detect embedding dimension: %w", err)
	}

	logger.Info("Ollama嵌入服务初始化成功",
		String("base_url", config.BaseURL),
		String("model", config.Model),
		Int("dimension", embedder.dimension))

	return embedder, nil
}

// Embed 批量嵌入多个文本
func (e *OllamaEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	e.logger.Debug("开始嵌入文本", Int("count", len(texts)))

	var allEmbeddings [][]float32

	// Ollama通常只支持单个文本嵌入，需要逐个处理
	for i, text := range texts {
		embedding, err := e.embedSingle(ctx, text)
		if err != nil {
			e.logger.Error("嵌入文本失败",
				Error(err),
				Int("index", i),
				String("text_preview", e.getTextPreview(text)))
			return nil, fmt.Errorf("failed to embed text at index %d: %w", i, err)
		}
		allEmbeddings = append(allEmbeddings, embedding)
	}

	e.logger.Debug("文本嵌入完成", Int("count", len(allEmbeddings)))
	return allEmbeddings, nil
}

// EmbedSingle 嵌入单个文本
func (e *OllamaEmbedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	return e.embedSingle(ctx, text)
}

// BatchEmbed 分批处理大量文本
func (e *OllamaEmbedder) BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
	if batchSize <= 0 {
		batchSize = len(texts)
	}

	var allEmbeddings [][]float32
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := e.Embed(ctx, batch)
		if err != nil {
			return nil, err
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// GetDimension 获取嵌入维度
func (e *OllamaEmbedder) GetDimension() int {
	return e.dimension
}

// GetModel 获取模型名称
func (e *OllamaEmbedder) GetModel() string {
	return e.model
}

// Health 健康检查
func (e *OllamaEmbedder) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/version", e.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer e.closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

// embedSingle 嵌入单个文本（私有方法）
func (e *OllamaEmbedder) embedSingle(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("%s/api/embeddings", e.baseURL)

	reqData := ollamaEmbedRequest{
		Model:  e.model,
		Prompt: text,
	}

	var respData ollamaEmbedResponse
	if err := e.makeRequest(ctx, url, reqData, &respData); err != nil {
		return nil, fmt.Errorf("failed to embed text: %w", err)
	}

	// 转换 []float64 到 []float32
	result := make([]float32, len(respData.Embedding))
	for i, val := range respData.Embedding {
		result[i] = float32(val)
	}

	return result, nil
}

// detectDimension 检测嵌入维度（私有方法）
func (e *OllamaEmbedder) detectDimension(ctx context.Context) error {
	// 使用测试文本获取嵌入维度
	embedding, err := e.embedSingle(ctx, "test")
	if err != nil {
		return err
	}

	e.dimension = len(embedding)
	e.logger.Debug("检测到嵌入维度", Int("dimension", e.dimension))
	return nil
}

// makeRequest 发送请求到Ollama（私有方法）
func (e *OllamaEmbedder) makeRequest(ctx context.Context, url string, reqData interface{}, respData interface{}) error {
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer e.closeResponse(resp)

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyBytes, respData)
}

// closeResponse 安全关闭响应体（私有方法）
func (e *OllamaEmbedder) closeResponse(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		e.logger.Warn("关闭响应体失败", Error(err))
	}
}

// getTextPreview 获取文本预览用于日志（私有方法）
func (e *OllamaEmbedder) getTextPreview(text string) string {
	if len(text) <= 50 {
		return text
	}
	return text[:47] + "..."
}