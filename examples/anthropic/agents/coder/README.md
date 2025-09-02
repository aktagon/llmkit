# Code-Editing Agent

A simple yet powerful code-editing agent that demonstrates how to build an AI assistant capable of reading, listing, and editing files using the llmkit framework.

## Features

- **File Reading**: Read the contents of any file in the working directory
- **Directory Listing**: List files and directories to explore the project structure  
- **File Editing**: Edit files using string replacement, or create new files from scratch
- **Interactive Chat**: Natural language interface for all file operations

## Tools Available

### `read_file`
Reads the contents of a specified file.
- **Input**: `path` (string) - relative path to the file
- **Example**: "What's in main.go?"

### `list_files` 
Lists files and directories in a given path.
- **Input**: `path` (string, optional) - defaults to current directory
- **Example**: "What files are in this directory?"

### `edit_file`
Edits files by replacing old text with new text, or creates new files.
- **Input**: 
  - `path` (string) - file path to edit
  - `old_str` (string) - text to replace (empty string to create new file)
  - `new_str` (string) - replacement text
- **Example**: "Create a calculator.py file that adds two numbers"

## Prerequisites

- Go 1.21 or later
- Anthropic API key

## Setup

1. Set your Anthropic API key:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

2. Run the agent:
   ```bash
   cd examples/anthropic/agents/coder
   go run main.go
   ```

## Usage Examples

```
=== Code-Editing Agent ===
I can help you read, list, and edit files!
Try: 'What files are in this directory?'
Or: 'Create a calculator.py file that adds two numbers'
Type 'exit' to quit.

You: What files are in this directory?
Agent: I'll check what files are in the current directory for you.

You: Create a simple calculator.py file that adds two numbers
Agent: I'll create a calculator.py file with a function to add two numbers...

You: Read the calculator.py file and explain what it does
Agent: I'll read the calculator.py file to see what it contains...

You: Edit the file to also include subtraction
Agent: I'll modify the calculator.py file to include a subtraction function...
```

## How It Works

This agent demonstrates the core concepts from the blog post "How to Build an Agent":

1. **Tool Registration**: Each file operation is implemented as a tool with a clear schema
2. **Conversation Loop**: The agent maintains context and handles tool execution automatically
3. **Natural Interface**: Users can request file operations in natural language

The agent uses Claude's built-in tool-use capabilities to determine when to read, list, or edit files based on user requests. All tool execution happens automatically - you just describe what you want to do with files and the agent figures out how to accomplish it.

## Architecture

- **main.go**: Core agent implementation with interactive chat loop
- **Tool Registration**: Each tool is registered with proper input schemas and handlers
- **Error Handling**: Comprehensive error handling for file operations
- **llmkit Integration**: Uses the llmkit agent framework for simplified development

This example shows how just a few hundred lines of Go code can create a surprisingly capable code-editing assistant.
---
Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com

