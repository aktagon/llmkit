# Workflow

A minimal Go library for creating and executing workflows with conditional task transitions.

## Features

- **Minimal API**: Just 1 interface and 3 methods to learn
- **Conditional Branching**: Route workflow execution based on task results
- **Context Support**: Built-in cancellation and timeout support
- **State Sharing**: Pass data between tasks using a shared state map
- **Loop Support**: Create workflows with cycles and retry logic
- **Zero Dependencies**: Uses only Go standard library

## Installation

```bash
go get github.com/aktagon/llmkit/workflow
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/aktagon/llmkit/workflow"
)

// Define a task
type GreetTask struct{}

func (t *GreetTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    name := state["name"].(string)
    state["message"] = fmt.Sprintf("Hello, %s!", name)
    return "default", nil
}

func main() {
    // Create workflow
    greet := &GreetTask{}
    wf := workflow.New(greet)

    // Execute workflow
    state := map[string]interface{}{"name": "World"}
    err := wf.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(state["message"]) // Output: Hello, World!
}
```

## API Reference

### Task Interface

```go
type Task interface {
    Execute(ctx context.Context, state map[string]interface{}) (string, error)
}
```

The `Task` interface represents a single unit of work. The `Execute` method:

- Receives a `context.Context` for cancellation and timeouts
- Receives a `state` map for reading and writing shared data
- Returns an `action` string that determines the next task
- Returns an `error` if the task fails

### Workflow Type

```go
type Workflow struct {
    // private fields
}
```

The `Workflow` type orchestrates task execution with conditional transitions.

#### Constructor

```go
func New(startTask Task) *Workflow
```

Creates a new workflow starting with the given task.

#### Methods

```go
func (w *Workflow) On(from Task, action string, to Task) *Workflow
```

Adds a transition from one task to another based on the action returned by the source task. This method is chainable.

```go
func (w *Workflow) Run(ctx context.Context, state map[string]interface{}) error
```

Executes the workflow starting from the start task. Continues until no more transitions are found or an error occurs.

## Examples

### Sequential Workflow

```go
type Task1 struct{}
func (t *Task1) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    state["step"] = 1
    return "next", nil
}

type Task2 struct{}
func (t *Task2) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    state["step"] = 2
    return "default", nil
}

func main() {
    task1 := &Task1{}
    task2 := &Task2{}

    wf := workflow.New(task1).On(task1, "next", task2)

    state := make(map[string]interface{})
    err := wf.Run(context.Background(), state)
    // state["step"] == 2
}
```

### Conditional Branching

```go
type DecisionTask struct{}
func (t *DecisionTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    score := state["score"].(int)
    if score >= 80 {
        return "approved", nil
    }
    return "rejected", nil
}

type ApprovalTask struct{}
func (t *ApprovalTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    state["status"] = "approved"
    return "default", nil
}

type RejectionTask struct{}
func (t *RejectionTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    state["status"] = "rejected"
    return "default", nil
}

func main() {
    decision := &DecisionTask{}
    approval := &ApprovalTask{}
    rejection := &RejectionTask{}

    wf := workflow.New(decision).
        On(decision, "approved", approval).
        On(decision, "rejected", rejection)

    state := map[string]interface{}{"score": 85}
    err := wf.Run(context.Background(), state)
    // state["status"] == "approved"
}
```

### Loop with Retry Logic

```go
type ProcessTask struct{}
func (t *ProcessTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    attempts := state["attempts"].(int)
    attempts++
    state["attempts"] = attempts

    // Simulate success after 3 attempts
    if attempts >= 3 {
        state["result"] = "success"
        return "complete", nil
    }

    return "retry", nil
}

type RetryTask struct{}
func (t *RetryTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    // Add delay logic here if needed
    time.Sleep(100 * time.Millisecond)
    return "continue", nil
}

type CompleteTask struct{}
func (t *CompleteTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    state["status"] = "completed"
    return "default", nil
}

func main() {
    process := &ProcessTask{}
    retry := &RetryTask{}
    complete := &CompleteTask{}

    wf := workflow.New(process).
        On(process, "retry", retry).
        On(process, "complete", complete).
        On(retry, "continue", process) // Loop back

    state := map[string]interface{}{"attempts": 0}
    err := wf.Run(context.Background(), state)
    // state["result"] == "success"
}
```

### Context Cancellation

```go
func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    wf := workflow.New(longRunningTask)

    state := make(map[string]interface{})
    err := wf.Run(ctx, state)
    if err == context.DeadlineExceeded {
        fmt.Println("Workflow timed out")
    }
}
```

## Patterns and Best Practices

### 1. State Management

Use descriptive keys and consistent types in the state map:

```go
func (t *Task) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    // Good: descriptive keys
    userID := state["user_id"].(string)
    orderTotal := state["order_total"].(float64)

    // Set results with clear naming
    state["payment_status"] = "completed"
    state["transaction_id"] = "txn_12345"

    return "success", nil
}
```

### 2. Action Naming

Use clear, descriptive action names:

```go
// Good: descriptive actions
return "payment_approved", nil
return "inventory_insufficient", nil
return "user_validation_failed", nil

// Avoid: generic actions
return "ok", nil
return "error", nil
return "next", nil
```

### 3. Error Handling

Handle errors explicitly and provide context:

```go
func (t *PaymentTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    amount := state["amount"].(float64)

    if amount <= 0 {
        return "", fmt.Errorf("invalid amount: %f", amount)
    }

    // Process payment
    err := processPayment(amount)
    if err != nil {
        return "", fmt.Errorf("payment processing failed: %w", err)
    }

    return "payment_success", nil
}
```

### 4. Context Awareness

Always check context cancellation in long-running tasks:

```go
func (t *LongTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    for i := 0; i < 1000; i++ {
        // Check for cancellation
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
        }

        // Do work
        processItem(i)
    }

    return "complete", nil
}
```

### 5. Task Composition

Keep tasks focused on a single responsibility:

```go
// Good: focused tasks
type ValidateOrderTask struct{}
type CalculateTaxTask struct{}
type ProcessPaymentTask struct{}

// Avoid: monolithic tasks
type HandleOrderTask struct{} // Does too much
```

### 6. Testing

Test tasks in isolation and workflows end-to-end:

```go
func TestValidateOrderTask(t *testing.T) {
    task := &ValidateOrderTask{}

    state := map[string]interface{}{
        "order_id": "12345",
        "items":    []string{"item1", "item2"},
    }

    action, err := task.Execute(context.Background(), state)

    assert.NoError(t, err)
    assert.Equal(t, "valid", action)
    assert.True(t, state["order_valid"].(bool))
}
```

## Common Use Cases

### Document Processing Pipeline

```go
validate := &ValidateDocumentTask{}
extract := &ExtractDataTask{}
transform := &TransformDataTask{}
store := &StoreDataTask{}
errorHandler := &ErrorHandlerTask{}

wf := workflow.New(validate).
    On(validate, "valid", extract).
    On(validate, "invalid", errorHandler).
    On(extract, "success", transform).
    On(extract, "failed", errorHandler).
    On(transform, "success", store).
    On(transform, "failed", errorHandler)
```

### User Registration Flow

```go
validateEmail := &ValidateEmailTask{}
checkExists := &CheckUserExistsTask{}
createUser := &CreateUserTask{}
sendWelcome := &SendWelcomeEmailTask{}
handleError := &HandleErrorTask{}

wf := workflow.New(validateEmail).
    On(validateEmail, "valid", checkExists).
    On(validateEmail, "invalid", handleError).
    On(checkExists, "new_user", createUser).
    On(checkExists, "exists", handleError).
    On(createUser, "created", sendWelcome)
```

### Order Processing System

```go
checkInventory := &CheckInventoryTask{}
processPayment := &ProcessPaymentTask{}
fulfillOrder := &FulfillOrderTask{}
sendConfirmation := &SendConfirmationTask{}
handleBackorder := &HandleBackorderTask{}
handlePaymentError := &HandlePaymentErrorTask{}

wf := workflow.New(checkInventory).
    On(checkInventory, "in_stock", processPayment).
    On(checkInventory, "out_of_stock", handleBackorder).
    On(processPayment, "payment_success", fulfillOrder).
    On(processPayment, "payment_failed", handlePaymentError).
    On(fulfillOrder, "fulfilled", sendConfirmation)
```

## Performance Considerations

- **Minimal Overhead**: Each transition has minimal overhead (~1μs)
- **Memory Efficient**: No unnecessary allocations during execution
- **Concurrent Safe**: Multiple workflows can run concurrently
- **Context Aware**: Respects cancellation and timeouts

## Limitations

- **No Parallel Execution**: Tasks execute sequentially
- **No Built-in Persistence**: State is not automatically persisted
- **No Built-in Retry Logic**: Must be implemented in tasks
- **No Built-in Logging**: Must be added to individual tasks

## Extensions

For advanced features, consider these patterns:

### Retry Logic

```go
type RetryableTask struct {
    maxRetries int
    task       Task
}

func (rt *RetryableTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    var err error
    for i := 0; i < rt.maxRetries; i++ {
        action, err := rt.task.Execute(ctx, state)
        if err == nil {
            return action, nil
        }

        // Wait before retry
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return "", err
}
```

### Logging

```go
type LoggingTask struct {
    task   Task
    logger *log.Logger
}

func (lt *LoggingTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
    lt.logger.Printf("Executing task: %T", lt.task)

    start := time.Now()
    action, err := lt.task.Execute(ctx, state)
    duration := time.Since(start)

    if err != nil {
        lt.logger.Printf("Task failed after %v: %v", duration, err)
    } else {
        lt.logger.Printf("Task completed in %v, action: %s", duration, action)
    }

    return action, err
}
```

## License

This package is part of the llmkit project and follows the same license terms.
