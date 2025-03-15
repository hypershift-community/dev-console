package args

import (
	"strings"

	"github.com/hypershift-community/dev-console/pkg/task"
	"github.com/hypershift-community/dev-console/pkg/task/taskfile/ast"
)

// Parse parses command line argument: tasks and global variables
func Parse(args ...string) ([]*task.Call, *ast.Vars) {
	calls := []*task.Call{}
	globals := ast.NewVars()

	for _, arg := range args {
		if !strings.Contains(arg, "=") {
			calls = append(calls, &task.Call{Task: arg})
			continue
		}

		name, value := splitVar(arg)
		globals.Set(name, ast.Var{Value: value})
	}

	return calls, globals
}

func splitVar(s string) (string, string) {
	pair := strings.SplitN(s, "=", 2)
	return pair[0], pair[1]
}
