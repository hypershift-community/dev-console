package task

import "github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"

// Call is the parameters to a task call
type Call struct {
	Task     string
	Vars     *ast.Vars
	Silent   bool
	Indirect bool // True if the task was called by another task
}
