// Package main demonstrates a script writing workflow using the workflow package
// and Anthropic API integration.
//
// This workflow:
// 1. Writes a script based on user description (limited to 1000 tokens)
// 2. Scores the script on a scale from 0-100
// 3. Judges the script based on "The Elements of Style" (Strunk & White)
// 4. Rewrites the script if score < 95
// 5. Publishes the final script to scripts/ directory
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/anthropic/types"
	"github.com/aktagon/llmkit/workflow"
)

// ScriptData holds the script content and metadata
type ScriptData struct {
	Content     string    `json:"content"`
	Score       int       `json:"score"`
	Feedback    string    `json:"feedback"`
	Attempts    int       `json:"attempts"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Spinner provides a visual progress indicator
type Spinner struct {
	mu      sync.Mutex
	active  bool
	message string
	done    chan bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan bool),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return
	}

	s.active = true
	go s.spin()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	s.done <- true
	fmt.Print("\r") // Clear the spinner line
}

func (s *Spinner) spin() {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for {
		select {
		case <-s.done:
			return
		default:
			fmt.Printf("\r%s %s", chars[i%len(chars)], s.message)
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

// TaskError wraps errors with task context
type TaskError struct {
	TaskName string
	Err      error
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("task '%s' failed: %v", e.TaskName, e.Err)
}

// extractJSON removes markdown code blocks and extracts JSON
func extractJSON(text string) string {
	// Remove markdown code blocks
	re := regexp.MustCompile("```(?:json)?\n?(.*?)\n?```")
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// If no code blocks, return the text as-is
	return strings.TrimSpace(text)
}

// callAnthropic makes a direct HTTP request to the Anthropic API
func callAnthropic(endpoint, apiKey string, requestBody string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", types.AnthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyText))
	}

	return string(bodyText), nil
}

// WriteScriptTask generates the initial script based on user description
type WriteScriptTask struct {
	APIKey string
}

func (t *WriteScriptTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	description := state["description"].(string)

	spinner := NewSpinner("Generating initial script...")
	spinner.Start()
	defer spinner.Stop()

	systemPrompt := `You are a professional screenwriter and script writer. You write engaging, well-structured scripts with compelling dialogue, clear scene descriptions, and proper formatting. Your scripts follow industry standards and best practices. Keep scripts concise but complete.`

	userPrompt := fmt.Sprintf(`Write a complete but concise script based on this description: %s

Requirements:
- Proper script formatting
- Scene headings (INT./EXT.)
- Character names in ALL CAPS when first introduced
- Clear dialogue and action lines
- Stage directions in parentheses when needed
- Keep it concise but complete (aim for 800-1000 tokens)

Make the script engaging and professionally formatted.`, description)

	// Build request manually to add max_tokens
	requestBody := map[string]interface{}{
		"model":      types.Model,
		"max_tokens": 1000, // Limit script length
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", &TaskError{"WriteScriptTask", fmt.Errorf("failed to marshal request: %w", err)}
	}

	response, err := callAnthropic(types.Endpoint, t.APIKey, string(jsonData))
	if err != nil {
		return "", &TaskError{"WriteScriptTask", fmt.Errorf("failed to generate script: %w", err)}
	}

	if len(response.Content) == 0 {
		return "", &TaskError{"WriteScriptTask", fmt.Errorf("no content in API response")}
	}

	scriptContent := response.Content[0].Text

	// Store the script in state
	state["script_content"] = scriptContent
	state["attempts"] = 1

	spinner.Stop()
	fmt.Printf("✓ Generated initial script (%d characters)\n", len(scriptContent))

	return "score", nil
}

// ScoreScriptTask scores the script on a scale from 0-100
type ScoreScriptTask struct {
	APIKey string
}

func (t *ScoreScriptTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	scriptContent := state["script_content"].(string)

	spinner := NewSpinner("Scoring script...")
	spinner.Start()
	defer spinner.Stop()

	systemPrompt := `You are a professional script editor and critic. You evaluate scripts based on:
- Story structure and pacing
- Character development and dialogue quality
- Scene construction and formatting
- Overall engagement and readability
- Professional writing standards

You provide scores from 0-100 where:
- 0-59: Poor quality, needs major revision
- 60-79: Average quality, needs improvement
- 80-94: Good quality, minor improvements needed
- 95-100: Excellent quality, publication ready

IMPORTANT: Respond with ONLY the JSON object, no additional text or formatting.`

	userPrompt := fmt.Sprintf(`Please score this script on a scale from 0-100 and provide brief feedback explaining your score:

SCRIPT:
%s

Respond with only this JSON structure:
{"score": <number>, "feedback": "<explanation>"}`, scriptContent)

	// Use structured output to get consistent scoring
	jsonSchema := `{
		"name": "script_score",
		"description": "Score and feedback for a script",
		"strict": true,
		"schema": {
			"type": "object",
			"properties": {
				"score": {
					"type": "integer",
					"minimum": 0,
					"maximum": 100,
					"description": "Numerical score from 0-100"
				},
				"feedback": {
					"type": "string",
					"description": "Brief explanation of the score and areas for improvement"
				}
			},
			"required": ["score", "feedback"],
			"additionalProperties": false
		}
	}`

	response, err := anthropic.Prompt(systemPrompt, userPrompt, jsonSchema, t.APIKey)
	if err != nil {
		return "", &TaskError{"ScoreScriptTask", fmt.Errorf("failed to score script: %w", err)}
	}

	if len(response.Content) == 0 {
		return "", &TaskError{"ScoreScriptTask", fmt.Errorf("no content in API response")}
	}

	// Extract and clean JSON content
	jsonContent := extractJSON(response.Content[0].Text)

	// Parse the structured JSON response
	var scoreResult struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &scoreResult); err != nil {
		return "", &TaskError{"ScoreScriptTask", fmt.Errorf("failed to parse score response (content: %q): %w", jsonContent, err)}
	}

	state["score"] = scoreResult.Score
	state["feedback"] = scoreResult.Feedback

	spinner.Stop()
	fmt.Printf("✓ Script scored: %d/100\n", scoreResult.Score)
	fmt.Printf("  Feedback: %s\n", scoreResult.Feedback)

	return "judge", nil
}

// JudgeScriptTask judges the script based on "The Elements of Style"
type JudgeScriptTask struct {
	APIKey string
}

func (t *JudgeScriptTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	scriptContent := state["script_content"].(string)
	score := state["score"].(int)

	spinner := NewSpinner("Judging script based on Elements of Style...")
	spinner.Start()
	defer spinner.Stop()

	systemPrompt := `You are a literary critic and editor who specializes in "The Elements of Style" by Strunk & White. You evaluate writing based on their core principles:

1. Use definite, specific, concrete language
2. Omit needless words
3. Use the active voice
4. Put statements in positive form
5. Choose a suitable design and hold to it
6. Make the paragraph the unit of composition
7. Use orthodox spelling

You provide detailed analysis of how well writing adheres to these principles and suggest specific improvements.`

	userPrompt := fmt.Sprintf(`Analyze this script based on "The Elements of Style" principles by Strunk & White. 

Current Score: %d/100

SCRIPT:
%s

Please provide:
1. Specific examples of where the script follows or violates Strunk & White principles
2. Concrete suggestions for improvement
3. Assessment of clarity, conciseness, and style
4. Overall recommendation for revision`, score, scriptContent)

	response, err := anthropic.Prompt(systemPrompt, userPrompt, "", t.APIKey)
	if err != nil {
		return "", &TaskError{"JudgeScriptTask", fmt.Errorf("failed to judge script: %w", err)}
	}

	if len(response.Content) == 0 {
		return "", &TaskError{"JudgeScriptTask", fmt.Errorf("no content in API response")}
	}

	judgment := response.Content[0].Text
	state["judgment"] = judgment

	spinner.Stop()
	fmt.Printf("✓ Script judged based on Elements of Style\n")

	// Decision based on score
	if score < 95 {
		fmt.Printf("  → Score %d < 95, proceeding to rewrite\n", score)
		return "rewrite", nil
	} else {
		fmt.Printf("  → Score %d >= 95, proceeding to publish\n", score)
		return "publish", nil
	}
}

// RewriteScriptTask rewrites the script based on feedback and judgment
type RewriteScriptTask struct {
	APIKey string
}

func (t *RewriteScriptTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	scriptContent := state["script_content"].(string)
	feedback := state["feedback"].(string)
	judgment := state["judgment"].(string)
	attempts := state["attempts"].(int)

	// Safety check to prevent infinite loops
	if attempts >= 5 {
		fmt.Printf("⚠ Maximum rewrite attempts reached (%d), proceeding to publish\n", attempts)
		return "publish", nil
	}

	spinner := NewSpinner(fmt.Sprintf("Rewriting script (attempt %d)...", attempts+1))
	spinner.Start()
	defer spinner.Stop()

	systemPrompt := `You are a professional script editor and rewriter. You improve scripts based on detailed feedback and literary criticism. You maintain the original story and intent while enhancing:

- Clarity and conciseness (Strunk & White principles)
- Dialogue quality and naturalness
- Scene structure and pacing
- Character development
- Professional formatting
- Overall readability and engagement

You make substantial improvements while preserving the script's core narrative. Keep the script concise (800-1000 tokens).`

	userPrompt := fmt.Sprintf(`Please rewrite this script based on the feedback and literary analysis provided. Make substantial improvements while maintaining the original story.

ORIGINAL SCRIPT:
%s

FEEDBACK:
%s

LITERARY ANALYSIS (Elements of Style):
%s

Please provide a significantly improved version that addresses all the issues mentioned in the feedback and analysis.`, scriptContent, feedback, judgment)

	// Build request manually to add max_tokens
	requestBody := map[string]interface{}{
		"model":      types.Model,
		"max_tokens": 1000, // Limit rewritten script length
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", &TaskError{"RewriteScriptTask", fmt.Errorf("failed to marshal request: %w", err)}
	}

	response, err := callAnthropic(types.Endpoint, t.APIKey, string(jsonData))
	if err != nil {
		return "", &TaskError{"RewriteScriptTask", fmt.Errorf("failed to rewrite script: %w", err)}
	}

	if len(response.Content) == 0 {
		return "", &TaskError{"RewriteScriptTask", fmt.Errorf("no content in API response")}
	}

	rewrittenScript := response.Content[0].Text

	// Update state with rewritten script
	state["script_content"] = rewrittenScript
	state["attempts"] = attempts + 1

	spinner.Stop()
	fmt.Printf("✓ Script rewritten (attempt %d)\n", attempts+1)

	// Go back to scoring the rewritten script
	return "score", nil
}

// PublishScriptTask saves the final script to the scripts/ directory
type PublishScriptTask struct{}

func (t *PublishScriptTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	scriptContent := state["script_content"].(string)
	description := state["description"].(string)
	score := state["score"].(int)
	attempts := state["attempts"].(int)

	spinner := NewSpinner("Publishing script...")
	spinner.Start()
	defer spinner.Stop()

	// Generate a filename based on description
	title := generateTitle(description)
	filename := fmt.Sprintf("%s_%d.txt", sanitizeFilename(title), time.Now().Unix())
	filepath := filepath.Join("scripts", filename)

	// Create script metadata
	scriptData := ScriptData{
		Content:     scriptContent,
		Score:       score,
		Feedback:    state["feedback"].(string),
		Attempts:    attempts,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}

	// Create the full script file with metadata
	fullContent := fmt.Sprintf(`# %s

**Description:** %s
**Score:** %d/100
**Attempts:** %d
**Created:** %s

---

%s

---

**Feedback:** %s
`, title, description, score, attempts, scriptData.CreatedAt.Format("2006-01-02 15:04:05"), scriptContent, state["feedback"].(string))

	// Ensure scripts directory exists
	if err := os.MkdirAll("scripts", 0755); err != nil {
		return "", &TaskError{"PublishScriptTask", fmt.Errorf("failed to create scripts directory: %w", err)}
	}

	// Write the script to file
	if err := os.WriteFile(filepath, []byte(fullContent), 0644); err != nil {
		return "", &TaskError{"PublishScriptTask", fmt.Errorf("failed to write script to file: %w", err)}
	}

	state["published_file"] = filepath
	state["title"] = title

	spinner.Stop()
	fmt.Printf("✓ Script published to: %s\n", filepath)
	fmt.Printf("  Title: %s\n", title)
	fmt.Printf("  Final Score: %d/100\n", score)
	fmt.Printf("  Total Attempts: %d\n", attempts)

	return "default", nil
}

// Helper functions

func generateTitle(description string) string {
	words := strings.Fields(description)
	if len(words) == 0 {
		return "Untitled Script"
	}

	// Take first few words as title
	maxWords := 4
	if len(words) < maxWords {
		maxWords = len(words)
	}

	title := strings.Join(words[:maxWords], " ")

	// Capitalize first letter
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	return title
}

func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return strings.ToLower(replacer.Replace(filename))
}

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Get script description from command line or use default
	description := "A short comedy sketch about two friends who accidentally switch phones and must navigate each other's lives for a day"
	if len(os.Args) > 1 {
		description = strings.Join(os.Args[1:], " ")
	}

	fmt.Printf("🎬 Starting Script Writing Workflow\n")
	fmt.Printf("Description: %s\n", description)
	fmt.Printf("Token Limit: 1000 per script\n\n")

	// Create tasks
	writeTask := &WriteScriptTask{APIKey: apiKey}
	scoreTask := &ScoreScriptTask{APIKey: apiKey}
	judgeTask := &JudgeScriptTask{APIKey: apiKey}
	rewriteTask := &RewriteScriptTask{APIKey: apiKey}
	publishTask := &PublishScriptTask{}

	// Build workflow with conditional logic
	wf := workflow.New(writeTask).
		On(writeTask, "score", scoreTask).
		On(scoreTask, "judge", judgeTask).
		On(judgeTask, "rewrite", rewriteTask).
		On(judgeTask, "publish", publishTask).
		On(rewriteTask, "score", scoreTask) // Loop back to scoring

	// Initialize state
	state := map[string]interface{}{
		"description": description,
	}

	// Execute workflow
	ctx := context.Background()
	start := time.Now()

	err := wf.Run(ctx, state)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	duration := time.Since(start)

	// Print final results
	fmt.Printf("\n🎉 Workflow completed successfully!\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Final file: %s\n", state["published_file"])
	fmt.Printf("Script title: %s\n", state["title"])
	fmt.Printf("Final score: %d/100\n", state["score"])
	fmt.Printf("Total attempts: %d\n", state["attempts"])
}
