package langchain

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"

	"github.com/jaeger-ai-assist-prototype/internal/ai"
)

type SearchExtractor struct {
	llm llms.Model
}

func NewSearchExtractor(model llms.Model) *SearchExtractor {
	return &SearchExtractor{llm: model}
}

// ---------- COMMON PROMPT EXECUTOR (ONE PLACE) ----------

func (e *SearchExtractor) generateWithPrompt(
	ctx context.Context,
	template string,
	context string,
) (string, error) {

	prompt := prompts.NewPromptTemplate(
		template,
		[]string{"Context"},
	)

	rendered, err := prompt.Format(map[string]any{
		"Context": context,
	})
	if err != nil {
		return "", err
	}

	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: rendered},
		},
	}

	resp, err := e.llm.GenerateContent(ctx, []llms.MessageContent{msg})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("LLM returned no choices")
	}

	raw := strings.TrimSpace(resp.Choices[0].Content)
	if raw == "" {
		return "", errors.New("LLM returned empty response")
	}

	return raw, nil
}

// ---------- JSON EXTRACTION HELPERS ----------

func parseJSONBlock(input string) string {
	re := regexp.MustCompile(`(?s)[\x60]{3}(?:json)?\s+(.*?)\s+[\x60]{3}`)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return strings.TrimSpace(input)
}

// ---------- FEATURE 1: NATURAL LANGUAGE → IR ----------

func (e *SearchExtractor) ExtractSearchIR(
	ctx context.Context,
	input string,
) (ai.SearchIR, error) {
	prompt := prompts.NewPromptTemplate(
		SearchExtractionPrompt,
		[]string{"Input"},
	)

	rendered, err := prompt.Format(map[string]any{
		"Input": input,
	})
	if err != nil {
		return ai.SearchIR{}, err
	}

	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: rendered},
		},
	}

	resp, err := e.llm.GenerateContent(ctx, []llms.MessageContent{msg})
	if err != nil {
		return ai.SearchIR{}, err
	}

	if len(resp.Choices) == 0 {
		return ai.SearchIR{}, errors.New("LLM returned no choices")
	}

	raw := strings.TrimSpace(resp.Choices[0].Content)
	if raw == "" {
		return ai.SearchIR{}, errors.New("LLM returned empty response")
	}

	cleanedJSON := parseJSONBlock(raw)
	if cleanedJSON == "" {
		return ai.SearchIR{}, errors.New("LLM returned empty or malformed response")
	}
	log.Println(cleanedJSON)

	var ir ai.SearchIR
	if err := json.Unmarshal([]byte(cleanedJSON), &ir); err != nil {
		return ai.SearchIR{}, errors.New("LLM output is not valid JSON")
	}

	return ir, nil
}

// ---------- FEATURE 2 & 3: EXPLAIN ----------

func (e *SearchExtractor) ExplainTrace(
	ctx context.Context,
	context string,
) (string, error) {
	log.Println(context)
	return e.generateWithPrompt(ctx, TraceExplainPrompt, context)
}

func (e *SearchExtractor) ExplainSpan(
	ctx context.Context,
	context string,
) (string, error) {
	log.Println(context)
	return e.generateWithPrompt(ctx, SpanExplainPrompt, context)
}
