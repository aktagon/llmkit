// Example program demonstrating the workflow package
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aktagon/llmkit/workflow"
)

// Example tasks
type StartTask struct{}

func (t *StartTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	fmt.Println("Starting workflow...")
	state["step"] = "started"
	return "proceed", nil
}

type ProcessingTask struct{}

func (t *ProcessingTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	fmt.Println("Processing...")
	state["step"] = "processing"
	state["data"] = "processed_data"
	return "complete", nil
}

type CompleteTask struct{}

func (t *CompleteTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
	fmt.Println("Workflow completed!")
	state["step"] = "completed"
	return "default", nil
}

func main() {
	// Create tasks
	start := &StartTask{}
	processing := &ProcessingTask{}
	complete := &CompleteTask{}

	// Build workflow
	wf := workflow.New(start).
		On(start, "proceed", processing).
		On(processing, "complete", complete)

	// Execute workflow
	state := make(map[string]interface{})
	err := wf.Run(context.Background(), state)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Print final state
	fmt.Printf("Final state: %+v\n", state)
}
