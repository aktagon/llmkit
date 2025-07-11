// Package workflow provides a minimal API for creating and executing workflows
// consisting of tasks with conditional transitions.
//
// A workflow is composed of tasks that implement the Task interface. Each task
// can read and modify shared state, and returns an action string that determines
// the next task to execute.
//
// Basic usage:
//
//	type MyTask struct{}
//
//	func (t *MyTask) Execute(ctx context.Context, state map[string]interface{}) (string, error) {
//		// Perform work
//		state["result"] = "completed"
//		return "success", nil
//	}
//
//	func main() {
//		task := &MyTask{}
//		wf := workflow.New(task)
//
//		state := make(map[string]interface{})
//		err := wf.Run(context.Background(), state)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// For more examples and patterns, see the examples_test.go file and the README.md.
package workflow
