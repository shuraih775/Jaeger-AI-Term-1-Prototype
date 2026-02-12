package langchain

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLM LLMConfig `yaml:"llm"`
}

type LLMConfig struct {
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Endpoint    string  `yaml:"endpoint"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
