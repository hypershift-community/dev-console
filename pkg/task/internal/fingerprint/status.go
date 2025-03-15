package fingerprint

import (
	"context"

	"github.com/hypershift-community/hyper-console/pkg/task/internal/env"
	"github.com/hypershift-community/hyper-console/pkg/task/internal/execext"
	"github.com/hypershift-community/hyper-console/pkg/task/internal/logger"
	"github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"
)

type StatusChecker struct {
	logger *logger.Logger
}

func NewStatusChecker(logger *logger.Logger) StatusCheckable {
	return &StatusChecker{
		logger: logger,
	}
}

func (checker *StatusChecker) IsUpToDate(ctx context.Context, t *ast.Task) (bool, error) {
	for _, s := range t.Status {
		err := execext.RunCommand(ctx, &execext.RunCommandOptions{
			Command: s,
			Dir:     t.Dir,
			Env:     env.Get(t),
		})
		if err != nil {
			checker.logger.VerboseOutf(logger.Yellow, "task: status command %s exited non-zero: %s\n", s, err)
			return false, nil
		}
		checker.logger.VerboseOutf(logger.Yellow, "task: status command %s exited zero\n", s)
	}
	return true, nil
}
