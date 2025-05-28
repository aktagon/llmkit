# Script Writing Workflow

This example demonstrates a sophisticated script writing workflow that uses the Anthropic API and the workflow package to create, evaluate, and refine scripts based on professional writing standards.

## Workflow Steps

The workflow implements the following process:

1. **Write Script** - Generate initial script based on user description
2. **Score Script** - Evaluate script quality on a scale of 0-100
3. **Judge Script** - Analyze script using "The Elements of Style" (Strunk & White) principles
4. **Rewrite Script** - Improve script if score < 95 (loops back to step 2)
5. **Publish Script** - Save final script to `scripts/` directory

## Features

- **Professional Script Generation**: Creates properly formatted scripts with scene headings, dialogue, and stage directions
- **Token-Limited Scripts**: Constrains script generation to 1000 tokens for concise, focused content
- **Objective Scoring**: Uses structured JSON output to get consistent 0-100 scores
- **Literary Analysis**: Applies "The Elements of Style" principles for quality assessment
- **Iterative Improvement**: Automatically rewrites scripts that score below 95
- **Loop Protection**: Limits rewrites to 5 attempts to prevent infinite loops
- **Rich Output**: Saves scripts with metadata, scores, and feedback
- **Progress Indicators**: Visual spinning progress indicators during API calls
- **Enhanced Error Handling**: Clear error messages with task names for easier debugging

## Requirements

- Go 1.24 or later
- Anthropic API key set as `ANTHROPIC_API_KEY` environment variable
- Internet connection for API calls

## Setup

1. Set your Anthropic API key:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

2. Navigate to the workflow directory:

```bash
cd examples/workflows/writing
```

## Usage

### Quick Start with Shell Script (Recommended)

Use the provided shell script for the easiest experience:

```bash
# Make the script executable (first time only)
chmod +x run.sh

# Run with default example
./run.sh

# Run with custom description
./run.sh "A romantic comedy about two baristas who compete for the best coffee recipe"
```

### Direct Go Execution

#### Default Example

Run with the built-in example description:

```bash
go run main.go
```

This will create a comedy sketch about friends who switch phones.

#### Custom Script Description

Provide your own script description:

```bash
go run main.go "A dramatic scene where a detective confronts the prime suspect in a murder case"
```

```bash
go run main.go "A romantic comedy about two baristas who compete for the best coffee recipe"
```

```bash
go run main.go "A thriller scene where someone discovers they're being followed through a dark parking garage"
```

### Test Mode

Run without API calls to test the workflow structure:

```bash
go run test.go
```

## Example Output

```
🎬 Starting Script Writing Workflow
Description: A short comedy sketch about two friends who accidentally switch phones
Token Limit: 1000 per script

⧗ Generating initial script...
✓ Generated initial script (947 characters)
⧗ Scoring script...
✓ Script scored: 78/100
  Feedback: Good dialogue and funny premise, but needs better pacing and more specific character development
⧗ Judging script based on Elements of Style...
✓ Script judged based on Elements of Style
  → Score 78 < 95, proceeding to rewrite
⧗ Rewriting script (attempt 2)...
✓ Script rewritten (attempt 2)
⧗ Scoring script...
✓ Script scored: 92/100
  Feedback: Much improved pacing and character development, minor dialogue refinements needed
⧗ Judging script based on Elements of Style...
✓ Script judged based on Elements of Style
  → Score 92 < 95, proceeding to rewrite
⧗ Rewriting script (attempt 3)...
✓ Script rewritten (attempt 3)
⧗ Scoring script...
✓ Script scored: 96/100
  Feedback: Excellent quality with natural dialogue and professional formatting
⧗ Judging script based on Elements of Style...
✓ Script judged based on Elements of Style
  → Score 96 >= 95, proceeding to publish
⧗ Publishing script...
✓ Script published to: scripts/a_short_comedy_sketch_1704067200.txt
  Title: A Short Comedy Sketch
  Final Score: 96/100
  Total Attempts: 3

🎉 Workflow completed successfully!
Duration: 32.1s
Final file: scripts/a_short_comedy_sketch_1704067200.txt
Final score: 96/100
Total attempts: 3
```

## Output Files

Scripts are saved in the `scripts/` directory with:

- **Filename**: `{title}_{timestamp}.txt`
- **Metadata**: Description, score, attempts, creation time
- **Content**: Full script with proper formatting
- **Feedback**: Final evaluation and suggestions

Example file structure:

```
# A Short Comedy Sketch

**Description:** A short comedy sketch about two friends who switch phones
**Score:** 96/100
**Attempts:** 3
**Created:** 2024-01-01 12:00:00

---

[SCRIPT CONTENT HERE]

---

**Feedback:** Excellent quality with natural dialogue and professional formatting
```

## Workflow Architecture

The example demonstrates several advanced workflow patterns:

### Conditional Branching

```go
wf := workflow.New(writeTask).
    On(writeTask, "score", scoreTask).
    On(scoreTask, "judge", judgeTask).
    On(judgeTask, "rewrite", rewriteTask).   // If score < 95
    On(judgeTask, "publish", publishTask).   // If score >= 95
    On(rewriteTask, "score", scoreTask)      // Loop back
```

### State Management

```go
state := map[string]interface{}{
    "description":    userInput,
    "script_content": generatedScript,
    "score":         scoreValue,
    "feedback":      feedbackText,
    "attempts":      attemptCount,
}
```

### API Integration

```go
response, err := anthropic.Prompt(systemPrompt, userPrompt, jsonSchema, apiKey)
```

### Structured Output

```go
jsonSchema := `{
    "name": "script_score",
    "description": "Score and feedback for a script",
    "strict": true,
    "schema": {
        "type": "object",
        "properties": {
            "score": {"type": "integer", "minimum": 0, "maximum": 100},
            "feedback": {"type": "string"}
        },
        "required": ["score", "feedback"]
    }
}`
```

## Customization

### Modify Scoring Criteria

Edit the `ScoreScriptTask` system prompt to change evaluation criteria:

```go
systemPrompt := `You are a professional script editor. Evaluate based on:
- Your custom criteria here
- Different scoring rubric
- Specific genre requirements`
```

### Change Quality Threshold

Modify the score threshold in `JudgeScriptTask`:

```go
if score < 90 {  // Changed from 95 to 90
    return "rewrite", nil
}
```

### Add New Tasks

Extend the workflow with additional tasks:

```go
formatTask := &FormatScriptTask{}
wf.On(judgeTask, "format", formatTask)
```

### Custom Output Directory

Change the output directory in `PublishScriptTask`:

```go
filepath := filepath.Join("my-scripts", filename)
```

## Error Handling

The workflow includes robust error handling:

- **API Failures**: Graceful error messages with context
- **Infinite Loops**: Maximum attempt limits (5 rewrites)
- **File System**: Directory creation and write permissions
- **JSON Parsing**: Structured output validation

## Performance Notes

- **API Calls**: Each workflow run makes 6-15 API calls depending on rewrites
- **Duration**: Typical runs take 20-40 seconds (improved with 1000-token limit)
- **Cost**: Estimate ~$0.05-0.15 per workflow run (reduced with token limits)
- **Token Usage**: Scripts limited to 1000 tokens each for efficiency

## Elements of Style Integration

The workflow specifically evaluates scripts based on Strunk & White's principles:

1. **Use definite, specific, concrete language**
2. **Omit needless words**
3. **Use the active voice**
4. **Put statements in positive form**
5. **Choose a suitable design and hold to it**
6. **Make the paragraph the unit of composition**
7. **Use orthodox spelling**

## Extensions

This example can be extended for:

- **Different Genres**: Adapt prompts for drama, thriller, documentary
- **Multiple Formats**: Screenplays, stage plays, radio scripts
- **Team Collaboration**: Multi-user feedback integration
- **Version Control**: Git integration for script history
- **Publishing Platforms**: Direct upload to script repositories

## Troubleshooting

### API Key Issues

```bash
# Check if API key is set
echo $ANTHROPIC_API_KEY

# Set temporarily
export ANTHROPIC_API_KEY="your-key"
```

### Permission Errors

```bash
# Ensure write permissions
chmod 755 scripts/
```

### Network Issues

- Check internet connection
- Verify API endpoint accessibility
- Consider timeout adjustments for slow connections

This example showcases the power of combining workflow orchestration with AI APIs to create sophisticated, iterative content creation processes.
