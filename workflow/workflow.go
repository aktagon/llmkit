// Package workflow provides a minimal API for creating and executing workflows
// consisting of tasks with conditional transitions.
package workflow

import (
	"context"
	"fmt"
	"reflect"
)

// Task represents a single unit of work in a workflow.
// The Execute method receives a context for cancellation and a state map for data sharing.
// It returns an action string that determines the next task, and an error if the task fails.
type Task interface {
	Execute(ctx context.Context, state map[string]interface{}) (string, error)
}

// Workflow orchestrates tasks with conditional transitions based on actions returned by tasks.
type Workflow struct {
	start       Task
	transitions map[Task]map[string]Task
}

// New creates a new workflow starting with the given task.
func New(startTask Task) *Workflow {
	return &Workflow{
		start:       startTask,
		transitions: make(map[Task]map[string]Task),
	}
}

// On adds a transition from one task to another based on the action returned by the from task.
// This method is chainable for convenience.
func (w *Workflow) On(from Task, action string, to Task) *Workflow {
	if w.transitions[from] == nil {
		w.transitions[from] = make(map[string]Task)
	}
	w.transitions[from][action] = to
	return w
}

// Run executes the workflow starting from the start task.
// It continues executing tasks based on the actions they return until no more transitions are found.
// The workflow can be cancelled using the context.
func (w *Workflow) Run(ctx context.Context, state map[string]interface{}) error {
	current := w.start

	for current != nil {
		// Check for context cancellation before executing task
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute current task
		action, err := current.Execute(ctx, state)
		if err != nil {
			// Wrap error with task information
			taskName := getTaskName(current)
			return fmt.Errorf("workflow task '%s' failed: %w", taskName, err)
		}

		// Find next task based on action
		if transitions, exists := w.transitions[current]; exists {
			current = transitions[action]
		} else {
			// No transition found, end workflow
			current = nil
		}
	}

	return nil
}

// getTaskName returns the name of a task using reflection
func getTaskName(task Task) string {
	if task == nil {
		return "<nil>"
	}

	// Get the type of the task
	t := reflect.TypeOf(task)

	// If it's a pointer, get the element type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}
