package llmkit

import (
	"context"
	"fmt"
)

// message represents a conversation message (internal type).
type message struct {
	role       string
	content    string
	toolCalls  []toolCall
	toolResult *toolResult
}

// toolCall represents a tool invocation (internal type).
type toolCall struct {
	id    string
	name  string
	input map[string]any
}

// toolResult represents a tool execution result (internal type).
type toolResult struct {
	toolUseID string
	content   string
}

// Agent manages a multi-turn conversation with tool support.
type Agent struct {
	provider Provider
	opts     *options
	tools    []Tool
	history  []message
	system   string
}

// NewAgent creates a new conversation agent.
func NewAgent(p Provider, opts ...Option) *Agent {
	return &Agent{
		provider: p,
		opts:     applyOptions(opts...),
		tools:    nil,
		history:  nil,
	}
}

// SetSystem sets the system prompt for the agent.
func (a *Agent) SetSystem(system string) {
	a.system = system
}

// AddTool registers a tool the agent can use.
func (a *Agent) AddTool(t Tool) {
	a.tools = append(a.tools, t)
}

// findTool returns the tool with the given name, or nil if not found.
func (a *Agent) findTool(name string) *Tool {
	for i := range a.tools {
		if a.tools[i].Name == name {
			return &a.tools[i]
		}
	}
	return nil
}

// Reset clears the conversation history and tools.
func (a *Agent) Reset() {
	a.history = nil
	a.tools = nil
}

// Chat sends a message and returns the response.
func (a *Agent) Chat(ctx context.Context, msg string) (Response, error) {
	// Add user message to history
	a.history = append(a.history, message{role: "user", content: msg})

	// If no tools registered, use simple path
	if len(a.tools) == 0 {
		return a.chatSimple(ctx)
	}

	// Tool loop
	return a.chatWithTools(ctx)
}

// chatSimple handles chat without tools.
func (a *Agent) chatSimple(ctx context.Context) (Response, error) {
	messages := make([]Message, len(a.history))
	for i, m := range a.history {
		messages[i] = Message{Role: m.role, Content: m.content}
	}

	req := Request{
		System:   a.system,
		Messages: messages,
	}

	resp, err := Prompt(ctx, a.provider, req, a.buildOpts()...)
	if err != nil {
		return Response{}, err
	}

	a.history = append(a.history, message{role: "assistant", content: resp.Text})
	return resp, nil
}

// chatWithTools handles chat with tool execution loop.
func (a *Agent) chatWithTools(ctx context.Context) (Response, error) {
	maxIter := a.opts.maxToolIterations
	if maxIter == 0 {
		maxIter = 10 // safety default
	}

	var totalUsage Usage

	for i := 0; i < maxIter; i++ {
		text, calls, usage, err := a.sendRequest(ctx)
		if err != nil {
			return Response{}, err
		}

		totalUsage.Input += usage.Input
		totalUsage.Output += usage.Output

		if len(calls) == 0 {
			// No tool calls - return final response
			a.history = append(a.history, message{role: "assistant", content: text})
			return Response{Text: text, Tokens: totalUsage}, nil
		}

		// Store assistant message with tool calls
		a.history = append(a.history, message{role: "assistant", toolCalls: calls})

		// Execute each tool
		for _, call := range calls {
			tool := a.findTool(call.name)
			if tool == nil {
				return Response{}, fmt.Errorf("unknown tool: %s", call.name)
			}

			result, err := tool.Run(call.input)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
			}

			a.history = append(a.history, message{
				role: "user",
				toolResult: &toolResult{
					toolUseID: call.id,
					content:   result,
				},
			})
		}
	}

	return Response{}, fmt.Errorf("exceeded max tool iterations (%d)", maxIter)
}

// sendRequest dispatches to the provider-specific tool function.
func (a *Agent) sendRequest(ctx context.Context) (string, []toolCall, Usage, error) {
	switch a.provider.Name {
	case Anthropic:
		return sendAnthropicWithTools(ctx, a.provider, a.history, a.system, a.tools, a.opts)
	case OpenAI, Grok:
		return sendOpenAIWithTools(ctx, a.provider, a.history, a.system, a.tools, a.opts)
	case Google:
		return sendGoogleWithTools(ctx, a.provider, a.history, a.system, a.tools, a.opts)
	default:
		return "", nil, Usage{}, fmt.Errorf("tool support not implemented for provider: %s", a.provider.Name)
	}
}

// ChatWithSchema sends a message and returns structured output.
func (a *Agent) ChatWithSchema(ctx context.Context, msg, schema string) (Response, error) {
	a.history = append(a.history, message{role: "user", content: msg})

	// Build messages from history
	messages := make([]Message, len(a.history))
	for i, m := range a.history {
		messages[i] = Message{Role: m.role, Content: m.content}
	}

	req := Request{
		System:   a.system,
		Messages: messages,
		Schema:   schema,
	}

	resp, err := Prompt(ctx, a.provider, req, a.buildOpts()...)
	if err != nil {
		return Response{}, err
	}

	a.history = append(a.history, message{role: "assistant", content: resp.Text})

	return resp, nil
}

// buildOpts returns options for the underlying Prompt call.
func (a *Agent) buildOpts() []Option {
	var opts []Option
	if a.opts.httpClient != nil {
		opts = append(opts, WithHTTPClient(a.opts.httpClient))
	}
	if a.opts.temperature != nil {
		opts = append(opts, WithTemperature(*a.opts.temperature))
	}
	if a.opts.maxTokens != nil {
		opts = append(opts, WithMaxTokens(*a.opts.maxTokens))
	}
	return opts
}
