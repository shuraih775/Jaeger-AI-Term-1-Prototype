package llm

import (
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"

	"github.com/jaeger-ai-assist-prototype/internal/llm/langchain"
)

func NewLLM(cfg langchain.LLMConfig) (llms.Model, error) {
	switch cfg.Provider {
	case "ollama":
		opts := []ollama.Option{
			ollama.WithModel(cfg.Model),
		}

		if cfg.Endpoint != "" {
			opts = append(opts, ollama.WithServerURL(cfg.Endpoint))
		}

		// if cfg.Temperature != 0 {
		// 	opts = append(opts, ollama.With(float32(cfg.Temperature)))
		// }

		// if cfg.MaxTokens > 0 {
		// 	opts = append(opts, ollama.WithPredictNumPredict(cfg.MaxTokens))
		// }

		opts = append(opts, ollama.WithPullModel())

		return ollama.New(opts...)

	default:
		return nil, errors.New("unsupported LLM provider")
	}
}
