package openai

// API Configuration Constants
const (
	// Model is the default OpenAI model to use
	Model = "gpt-4o-2024-08-06"

	// EndpointResponses is the OpenAI API endpoint for structured responses
	EndpointResponses = "https://api.openai.com/v1/responses"

	// EndpointCompletions is the OpenAI API endpoint for chat completions
	EndpointCompletions = "https://api.openai.com/v1/chat/completions"

	// EndpointSpeech is the OpenAI API endpoint for text-to-speech
	EndpointSpeech = "https://api.openai.com/v1/audio/speech"

	// EndpointTranscriptions is the OpenAI API endpoint for speech-to-text
	EndpointTranscriptions = "https://api.openai.com/v1/audio/transcriptions"

	// ModelTTS1 is optimized for real-time use cases
	ModelTTS1 = "tts-1"

	// ModelTTS1HD is optimized for quality
	ModelTTS1HD = "tts-1-hd"

	// ModelWhisper1 is the Whisper V2 model for transcription
	ModelWhisper1 = "whisper-1"

	// ModelGPT4OTranscribe is the GPT-4o model for transcription
	ModelGPT4OTranscribe = "gpt-4o-transcribe"

	// ModelGPT4OMiniTranscribe is the GPT-4o mini model for transcription
	ModelGPT4OMiniTranscribe = "gpt-4o-mini-transcribe"
)

// ToolHandler executes tool logic and returns results
type ToolHandler func(input map[string]interface{}) (string, error)

// Tool represents a tool/function that can be called by GPT
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
	Handler     ToolHandler `json:"-"`
}

// Message represents a conversation message in OpenAI format
type Message struct {
	Role         string        `json:"role"`
	Content      string        `json:"content,omitempty"`
	Name         string        `json:"name,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// FunctionCall represents a function call request from GPT
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Function represents a function in OpenAI format (for request)
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// Response represents the API response structure
type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role         string        `json:"role"`
			Content      string        `json:"content"`
			FunctionCall *FunctionCall `json:"function_call,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// JsonSchema represents the JSON schema format for structured output
type JsonSchema struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Schema interface{} `json:"schema"`
	Strict bool        `json:"strict"`
}

// SchemaValidation represents the expected schema structure
type SchemaValidation struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Strict      bool        `json:"strict"`
	Schema      interface{} `json:"schema"`
}

// TextFormat represents the text format configuration
type TextFormat struct {
	Format JsonSchema `json:"format"`
}

// ChatRequest represents a standard chat completion request
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// StructuredRequest represents a structured output request
type StructuredRequest struct {
	Model string     `json:"model"`
	Input []Message  `json:"input"`
	Text  TextFormat `json:"text"`
}

// ToolsRequest represents a request using the newer tools API
type ToolsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Tools []struct {
		Type        string      `json:"type"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Parameters  interface{} `json:"parameters"`
	} `json:"tools"`
}

// FunctionCallResponse represents a function call in the response
type FunctionCallResponse struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Voice represents available TTS voices
type Voice string

const (
	VoiceAlloy   Voice = "alloy"
	VoiceEcho    Voice = "echo"
	VoiceFable   Voice = "fable"
	VoiceNova    Voice = "nova"
	VoiceOnyx    Voice = "onyx"
	VoiceShimmer Voice = "shimmer"
)

// AudioFormat represents available audio response formats
type AudioFormat string

const (
	FormatMP3  AudioFormat = "mp3"
	FormatOpus AudioFormat = "opus"
	FormatAAC  AudioFormat = "aac"
	FormatFLAC AudioFormat = "flac"
	FormatWAV  AudioFormat = "wav"
	FormatPCM  AudioFormat = "pcm"
)

// TTSRequest represents a text-to-speech request
type TTSRequest struct {
	Model          string      `json:"model"`
	Input          string      `json:"input"`
	Voice          Voice       `json:"voice"`
	ResponseFormat AudioFormat `json:"response_format,omitempty"`
	Speed          float64     `json:"speed,omitempty"`
}

// TTSOptions provides optional parameters for TTS requests
type TTSOptions struct {
	Model          string
	Voice          Voice
	ResponseFormat AudioFormat
	Speed          float64
}

// STTResponseFormat represents available transcription response formats
type STTResponseFormat string

const (
	STTFormatJSON        STTResponseFormat = "json"
	STTFormatText        STTResponseFormat = "text"
	STTFormatSRT         STTResponseFormat = "srt"
	STTFormatVerboseJSON STTResponseFormat = "verbose_json"
	STTFormatVTT         STTResponseFormat = "vtt"
)

// TimestampGranularity represents timestamp granularity options
type TimestampGranularity string

const (
	GranularityWord    TimestampGranularity = "word"
	GranularitySegment TimestampGranularity = "segment"
)

// STTOptions provides optional parameters for STT requests
type STTOptions struct {
	Model                  string
	Language               string
	Prompt                 string
	ResponseFormat         STTResponseFormat
	Temperature            float64
	TimestampGranularities []TimestampGranularity
}

// STTResponse represents the transcription response
type STTResponse struct {
	Text     string       `json:"text"`
	Language string       `json:"language,omitempty"`
	Duration float64      `json:"duration,omitempty"`
	Words    []STTWord    `json:"words,omitempty"`
	Segments []STTSegment `json:"segments,omitempty"`
}

// STTWord represents a word in the transcription with timestamp
type STTWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// STTSegment represents a segment in the transcription with timestamp
type STTSegment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}
