package task

import (
	"cmp"
	"fmt"

	"github.com/hypershift-community/hyper-console/pkg/task/internal/hash"
	"github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"
)

func (e *Executor) GetHash(t *ast.Task) (string, error) {
	r := cmp.Or(t.Run, e.Taskfile.Run)
	var h hash.HashFunc
	switch r {
	case "always":
		h = hash.Empty
	case "once":
		h = hash.Name
	case "when_changed":
		h = hash.Hash
	default:
		return "", fmt.Errorf(`task: invalid run "%s"`, r)
	}
	return h(t)
}
