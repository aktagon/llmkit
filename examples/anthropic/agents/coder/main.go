package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/anthropic/agents"
)

// registerReadFileTool adds file reading capability
func registerReadFileTool(agent *agents.ChatAgent) error {
	tool := anthropic.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file at the given path. Use this when you want to see what's inside a file. Do not use this with directory names.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The relative path to the file to read",
				},
			},
			"required":             []string{"path"},
			"additionalProperties": false,
		},
		Handler: func(input map[string]interface{}) (string, error) {
			path, ok := input["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path is required and must be a string")
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", path, err)
			}

			return string(content), nil
		},
	}

	return agent.RegisterTool(tool)
}

// registerListFilesTool adds file listing capability
func registerListFilesTool(agent *agents.ChatAgent) error {
	tool := anthropic.Tool{
		Name:        "list_files",
		Description: "List files and directories in a given path. If no path is provided, lists files in the current directory.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Optional relative path to list files from. Defaults to current directory if not provided.",
				},
			},
			"additionalProperties": false,
		},
		Handler: func(input map[string]interface{}) (string, error) {
			dir := "."
			if pathInput, exists := input["path"]; exists {
				if pathStr, ok := pathInput.(string); ok && pathStr != "" {
					dir = pathStr
				}
			}

			var files []string
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(dir, path)
				if err != nil {
					return err
				}

				if relPath != "." {
					if info.IsDir() {
						files = append(files, relPath+"/")
					} else {
						files = append(files, relPath)
					}
				}
				return nil
			})

			if err != nil {
				return "", fmt.Errorf("failed to list files in %s: %w", dir, err)
			}

			if len(files) == 0 {
				return "No files found in directory", nil
			}

			return fmt.Sprintf("Files in %s:\n%s", dir, strings.Join(files, "\n")), nil
		},
	}

	return agent.RegisterTool(tool)
}

// registerEditFileTool adds file editing capability via string replacement
func registerEditFileTool(agent *agents.ChatAgent) error {
	tool := anthropic.Tool{
		Name: "edit_file",
		Description: `Edit a text file by replacing old text with new text.

Replaces 'old_str' with 'new_str' in the given file. The old_str and new_str must be different.

If the file doesn't exist and old_str is empty, creates a new file with new_str as content.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to edit",
				},
				"old_str": map[string]interface{}{
					"type":        "string",
					"description": "Text to search for and replace. Must match exactly. Use empty string to create new file.",
				},
				"new_str": map[string]interface{}{
					"type":        "string",
					"description": "Text to replace old_str with",
				},
			},
			"required":             []string{"path", "old_str", "new_str"},
			"additionalProperties": false,
		},
		Handler: func(input map[string]interface{}) (string, error) {
			path, ok := input["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path is required and must be a string")
			}

			oldStr, ok := input["old_str"].(string)
			if !ok {
				return "", fmt.Errorf("old_str is required and must be a string")
			}

			newStr, ok := input["new_str"].(string)
			if !ok {
				return "", fmt.Errorf("new_str is required and must be a string")
			}

			if oldStr == newStr {
				return "", fmt.Errorf("old_str and new_str must be different")
			}

			content, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) && oldStr == "" {
					return createNewFile(path, newStr)
				}
				return "", fmt.Errorf("failed to read file %s: %w", path, err)
			}

			oldContent := string(content)
			newContent := strings.ReplaceAll(oldContent, oldStr, newStr)

			if oldContent == newContent && oldStr != "" {
				return "", fmt.Errorf("old_str not found in file")
			}

			err = os.WriteFile(path, []byte(newContent), 0644)
			if err != nil {
				return "", fmt.Errorf("failed to write file %s: %w", path, err)
			}

			return "File edited successfully", nil
		},
	}

	return agent.RegisterTool(tool)
}

// createNewFile creates a new file with the given content
func createNewFile(filePath, content string) (string, error) {
	dir := filepath.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	agent, err := agents.New(apiKey)
	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	// Register file operation tools
	if err := registerReadFileTool(agent); err != nil {
		log.Fatalf("Failed to register read_file tool: %v", err)
	}

	if err := registerListFilesTool(agent); err != nil {
		log.Fatalf("Failed to register list_files tool: %v", err)
	}

	if err := registerEditFileTool(agent); err != nil {
		log.Fatalf("Failed to register edit_file tool: %v", err)
	}

	fmt.Println("=== Code-Editing Agent ===")
	fmt.Println("I can help you read, list, and edit files!")
	fmt.Println("Try: 'What files are in this directory?'")
	fmt.Println("Or: 'Create a hello.py file that prints Hello World'")
	fmt.Println("Type 'exit' to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "" {
			continue
		}

		response, err := agent.Chat(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\u001b[93mAgent\u001b[0m: %s\n\n", response.Text)
	}
}