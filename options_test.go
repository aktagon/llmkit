package llmkit

import (
	"context"
	"net/http"
	"testing"
)

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	opt := WithHTTPClient(custom)

	opts := &options{}
	opt(opts)

	if opts.httpClient != custom {
		t.Error("WithHTTPClient did not set httpClient")
	}
}

func TestWithBeforeRequest(t *testing.T) {
	called := false
	fn := func(ctx context.Context, req *Request) error {
		called = true
		return nil
	}

	opt := WithBeforeRequest(fn)
	opts := &options{}
	opt(opts)

	if opts.beforeRequest == nil {
		t.Fatal("WithBeforeRequest did not set beforeRequest")
	}

	opts.beforeRequest(context.Background(), &Request{})
	if !called {
		t.Error("beforeRequest was not called")
	}
}

func TestWithAfterResponse(t *testing.T) {
	called := false
	fn := func(ctx context.Context, resp *Response, err error) {
		called = true
	}

	opt := WithAfterResponse(fn)
	opts := &options{}
	opt(opts)

	if opts.afterResponse == nil {
		t.Fatal("WithAfterResponse did not set afterResponse")
	}

	opts.afterResponse(context.Background(), &Response{}, nil)
	if !called {
		t.Error("afterResponse was not called")
	}
}

func TestApplyOptions(t *testing.T) {
	custom := &http.Client{}

	opts := applyOptions(
		WithHTTPClient(custom),
	)

	if opts.httpClient != custom {
		t.Error("applyOptions did not apply WithHTTPClient")
	}
}

func TestApplyOptions_Defaults(t *testing.T) {
	opts := applyOptions()

	if opts.httpClient != http.DefaultClient {
		t.Error("default httpClient should be http.DefaultClient")
	}
}

func TestWithTemperature(t *testing.T) {
	opt := WithTemperature(0.7)
	opts := &options{}
	opt(opts)

	if opts.temperature == nil {
		t.Fatal("WithTemperature did not set temperature")
	}
	if *opts.temperature != 0.7 {
		t.Errorf("temperature = %v, want 0.7", *opts.temperature)
	}
}

func TestWithTopP(t *testing.T) {
	opt := WithTopP(0.9)
	opts := &options{}
	opt(opts)

	if opts.topP == nil {
		t.Fatal("WithTopP did not set topP")
	}
	if *opts.topP != 0.9 {
		t.Errorf("topP = %v, want 0.9", *opts.topP)
	}
}

func TestWithTopK(t *testing.T) {
	opt := WithTopK(40)
	opts := &options{}
	opt(opts)

	if opts.topK == nil {
		t.Fatal("WithTopK did not set topK")
	}
	if *opts.topK != 40 {
		t.Errorf("topK = %v, want 40", *opts.topK)
	}
}

func TestWithMaxTokens(t *testing.T) {
	opt := WithMaxTokens(4096)
	opts := &options{}
	opt(opts)

	if opts.maxTokens == nil {
		t.Fatal("WithMaxTokens did not set maxTokens")
	}
	if *opts.maxTokens != 4096 {
		t.Errorf("maxTokens = %v, want 4096", *opts.maxTokens)
	}
}

func TestWithStopSequences(t *testing.T) {
	opt := WithStopSequences("STOP", "END")
	opts := &options{}
	opt(opts)

	if len(opts.stopSequences) != 2 {
		t.Fatalf("stopSequences length = %d, want 2", len(opts.stopSequences))
	}
	if opts.stopSequences[0] != "STOP" || opts.stopSequences[1] != "END" {
		t.Errorf("stopSequences = %v, want [STOP END]", opts.stopSequences)
	}
}

func TestWithSeed(t *testing.T) {
	opt := WithSeed(12345)
	opts := &options{}
	opt(opts)

	if opts.seed == nil {
		t.Fatal("WithSeed did not set seed")
	}
	if *opts.seed != 12345 {
		t.Errorf("seed = %v, want 12345", *opts.seed)
	}
}

func TestWithFrequencyPenalty(t *testing.T) {
	opt := WithFrequencyPenalty(0.5)
	opts := &options{}
	opt(opts)

	if opts.frequencyPenalty == nil {
		t.Fatal("WithFrequencyPenalty did not set frequencyPenalty")
	}
	if *opts.frequencyPenalty != 0.5 {
		t.Errorf("frequencyPenalty = %v, want 0.5", *opts.frequencyPenalty)
	}
}

func TestWithPresencePenalty(t *testing.T) {
	opt := WithPresencePenalty(0.3)
	opts := &options{}
	opt(opts)

	if opts.presencePenalty == nil {
		t.Fatal("WithPresencePenalty did not set presencePenalty")
	}
	if *opts.presencePenalty != 0.3 {
		t.Errorf("presencePenalty = %v, want 0.3", *opts.presencePenalty)
	}
}

func TestApplyOptions_Override(t *testing.T) {
	opts := applyOptions(
		WithTemperature(0.5),
		WithMaxTokens(2048),
	)

	if opts.temperature == nil || *opts.temperature != 0.5 {
		t.Error("WithTemperature not applied")
	}
	if opts.maxTokens == nil || *opts.maxTokens != 2048 {
		t.Error("WithMaxTokens not applied")
	}
}

func TestWithThinkingBudget(t *testing.T) {
	opt := WithThinkingBudget(4096)
	opts := &options{}
	opt(opts)

	if opts.thinkingBudget == nil {
		t.Fatal("WithThinkingBudget did not set thinkingBudget")
	}
	if *opts.thinkingBudget != 4096 {
		t.Errorf("thinkingBudget = %v, want 4096", *opts.thinkingBudget)
	}
}

func TestWithReasoningEffort(t *testing.T) {
	opt := WithReasoningEffort("high")
	opts := &options{}
	opt(opts)

	if opts.reasoningEffort == "" {
		t.Fatal("WithReasoningEffort did not set reasoningEffort")
	}
	if opts.reasoningEffort != "high" {
		t.Errorf("reasoningEffort = %v, want high", opts.reasoningEffort)
	}
}
