package llmkit

import (
	"os"
	"strconv"
)

// defaults holds environment-configured defaults for generation parameters.
var defaults struct {
	temperature      *float64
	topP             *float64
	topK             *int
	maxTokens        *int
	seed             *int64
	frequencyPenalty *float64
	presencePenalty  *float64
	thinkingBudget   *int
	reasoningEffort  string
}

func init() {
	defaults.temperature = parseFloat("LLMKIT_TEMPERATURE")
	defaults.topP = parseFloat("LLMKIT_TOP_P")
	defaults.topK = parseInt("LLMKIT_TOP_K")
	defaults.maxTokens = parseInt("LLMKIT_MAX_TOKENS")
	defaults.seed = parseInt64("LLMKIT_SEED")
	defaults.frequencyPenalty = parseFloat("LLMKIT_FREQUENCY_PENALTY")
	defaults.presencePenalty = parseFloat("LLMKIT_PRESENCE_PENALTY")
	defaults.thinkingBudget = parseInt("LLMKIT_THINKING_BUDGET")
	defaults.reasoningEffort = parseString("LLMKIT_REASONING_EFFORT")
}

func parseFloat(key string) *float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	}
	return nil
}

func parseInt(key string) *int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return &i
		}
	}
	return nil
}

func parseInt64(key string) *int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return &i
		}
	}
	return nil
}

func parseString(key string) string {
	return os.Getenv(key)
}
