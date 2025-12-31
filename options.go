package llmkit

import (
	"context"
	"net/http"
)

// Option configures Prompt and Agent behavior.
type Option func(*options)

type options struct {
	httpClient    *http.Client
	beforeRequest func(ctx context.Context, req *Request) error
	afterResponse func(ctx context.Context, resp *Response, err error)

	// Generation parameters
	temperature      *float64
	topP             *float64
	topK             *int
	maxTokens        *int
	stopSequences    []string
	seed             *int64
	frequencyPenalty *float64
	presencePenalty  *float64
	thinkingBudget   *int
	reasoningEffort  string

	// Agent parameters
	maxToolIterations int
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) {
		o.httpClient = c
	}
}

// WithBeforeRequest sets a hook called before each request.
func WithBeforeRequest(fn func(ctx context.Context, req *Request) error) Option {
	return func(o *options) {
		o.beforeRequest = fn
	}
}

// WithAfterResponse sets a hook called after each response.
func WithAfterResponse(fn func(ctx context.Context, resp *Response, err error)) Option {
	return func(o *options) {
		o.afterResponse = fn
	}
}

// WithTemperature sets the sampling temperature (0.0-2.0).
func WithTemperature(v float64) Option {
	return func(o *options) {
		o.temperature = &v
	}
}

// WithTopP sets nucleus sampling threshold (0.0-1.0).
func WithTopP(v float64) Option {
	return func(o *options) {
		o.topP = &v
	}
}

// WithTopK limits sampling to top K tokens.
func WithTopK(n int) Option {
	return func(o *options) {
		o.topK = &n
	}
}

// WithMaxTokens sets the maximum tokens to generate.
func WithMaxTokens(n int) Option {
	return func(o *options) {
		o.maxTokens = &n
	}
}

// WithStopSequences sets strings that halt generation.
func WithStopSequences(s ...string) Option {
	return func(o *options) {
		o.stopSequences = s
	}
}

// WithSeed sets seed for deterministic generation (beta).
func WithSeed(n int64) Option {
	return func(o *options) {
		o.seed = &n
	}
}

// WithFrequencyPenalty penalizes token repetition (-2.0 to 2.0). OpenAI/Grok only.
func WithFrequencyPenalty(v float64) Option {
	return func(o *options) {
		o.frequencyPenalty = &v
	}
}

// WithPresencePenalty encourages topic diversity (-2.0 to 2.0). OpenAI/Grok only.
func WithPresencePenalty(v float64) Option {
	return func(o *options) {
		o.presencePenalty = &v
	}
}

// WithThinkingBudget sets the token budget for extended thinking. Anthropic and Google Gemini 2.5 only.
// Minimum 1024 tokens for Anthropic. Budget counts towards max_tokens limit.
func WithThinkingBudget(n int) Option {
	return func(o *options) {
		o.thinkingBudget = &n
	}
}

// WithReasoningEffort controls reasoning intensity ("low", "medium", "high"). OpenAI o-series and Google Gemini 3 only.
func WithReasoningEffort(v string) Option {
	return func(o *options) {
		o.reasoningEffort = v
	}
}

// WithMaxToolIterations sets the maximum tool execution iterations for Agent.Chat().
// Default is 10. Set to 0 for unlimited (use with caution).
func WithMaxToolIterations(n int) Option {
	return func(o *options) {
		o.maxToolIterations = n
	}
}

// applyOptions creates options with defaults and applies all provided options.
func applyOptions(opts ...Option) *options {
	o := &options{
		httpClient:        http.DefaultClient,
		temperature:       defaults.temperature,
		topP:              defaults.topP,
		topK:              defaults.topK,
		maxTokens:         defaults.maxTokens,
		seed:              defaults.seed,
		frequencyPenalty:  defaults.frequencyPenalty,
		presencePenalty:   defaults.presencePenalty,
		thinkingBudget:    defaults.thinkingBudget,
		reasoningEffort:   defaults.reasoningEffort,
		maxToolIterations: 10,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
