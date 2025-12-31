package llmkit

import "testing"

func TestProviderBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		path     string
		want     string
	}{
		{
			name:     "openai default",
			provider: Provider{Name: OpenAI, APIKey: "key"},
			path:     "/v1/chat/completions",
			want:     "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "openai custom base url",
			provider: Provider{Name: OpenAI, APIKey: "key", BaseURL: "http://localhost:4000"},
			path:     "/v1/chat/completions",
			want:     "http://localhost:4000/v1/chat/completions",
		},
		{
			name:     "anthropic default",
			provider: Provider{Name: Anthropic, APIKey: "key"},
			path:     "/v1/messages",
			want:     "https://api.anthropic.com/v1/messages",
		},
		{
			name:     "anthropic custom base url",
			provider: Provider{Name: Anthropic, APIKey: "key", BaseURL: "https://proxy.example.com"},
			path:     "/v1/messages",
			want:     "https://proxy.example.com/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.provider.buildURL(tt.path)
			if got != tt.want {
				t.Errorf("Provider.buildURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProviderModel(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		want     string
	}{
		{
			name:     "anthropic default",
			provider: Provider{Name: Anthropic, APIKey: "key"},
			want:     "claude-sonnet-4-5",
		},
		{
			name:     "openai default",
			provider: Provider{Name: OpenAI, APIKey: "key"},
			want:     "gpt-4o-2024-08-06",
		},
		{
			name:     "google default",
			provider: Provider{Name: Google, APIKey: "key"},
			want:     "gemini-2.5-flash",
		},
		{
			name:     "grok default",
			provider: Provider{Name: Grok, APIKey: "key"},
			want:     "grok-3-fast",
		},
		{
			name:     "custom model override",
			provider: Provider{Name: Anthropic, APIKey: "key", Model: "claude-3-opus"},
			want:     "claude-3-opus",
		},
		{
			name:     "unknown provider returns empty",
			provider: Provider{Name: "unknown", APIKey: "key"},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.provider.model()
			if got != tt.want {
				t.Errorf("Provider.model() = %q, want %q", got, tt.want)
			}
		})
	}
}
