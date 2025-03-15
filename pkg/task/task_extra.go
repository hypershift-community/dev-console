/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package task

import (
	"context"
	"sync/atomic"

	"mvdan.cc/sh/v3/interp"

	"github.com/hypershift-community/hyper-console/pkg/task/errors"
	"github.com/hypershift-community/hyper-console/pkg/task/internal/fingerprint"
	"github.com/hypershift-community/hyper-console/pkg/task/internal/logger"
	"github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"
)

func (e *Executor) PrepareTask(call *Call) (*ast.Task, error) {
	t, err := e.FastCompiledTask(call)
	if err != nil {
		return nil, err
	}
	if !shouldRunOnCurrentPlatform(t.Platforms) {
		e.Logger.VerboseOutf(logger.Yellow, `task: %q not for current platform - ignored\n`, call.Task)
		return nil, nil
	}

	if err := e.areTaskRequiredVarsSet(t); err != nil {
		return nil, err
	}

	t, err = e.CompiledTask(call)
	if err != nil {
		return nil, err
	}

	if err := e.areTaskRequiredVarsAllowedValuesSet(t); err != nil {
		return nil, err
	}
	return t, nil
}

// RunTaskCmd runs a task by its name
func (e *Executor) RunTaskCmd(ctx context.Context, call *Call, t *ast.Task, cmdIndex int) error {
	if cmdIndex < 0 || cmdIndex >= len(t.Cmds) {
		return &errors.TaskCmdIndexError{
			TaskName: t.Task,
			CmdIndex: cmdIndex,
		}
	}

	if !e.Watch && atomic.AddInt32(e.taskCallCount[t.Task], 1) >= MaximumTaskCall {
		return &errors.TaskCalledTooManyTimesError{
			TaskName:        t.Task,
			MaximumTaskCall: MaximumTaskCall,
		}
	}

	release := e.acquireConcurrencyLimit()
	defer release()

	return e.startExecution(ctx, t, func(ctx context.Context) error {
		e.Logger.VerboseErrf(logger.Magenta, "task: %q started\n", call.Task)
		if err := e.runDeps(ctx, t); err != nil {
			return err
		}

		skipFingerprinting := e.ForceAll || (!call.Indirect && e.Force)
		if !skipFingerprinting {
			if err := ctx.Err(); err != nil {
				return err
			}

			preCondMet, err := e.areTaskPreconditionsMet(ctx, t)
			if err != nil {
				return err
			}

			// Get the fingerprinting method to use
			method := e.Taskfile.Method
			if t.Method != "" {
				method = t.Method
			}

			upToDate, err := fingerprint.IsTaskUpToDate(ctx, t,
				fingerprint.WithMethod(method),
				fingerprint.WithTempDir(e.TempDir.Fingerprint),
				fingerprint.WithDry(e.Dry),
				fingerprint.WithLogger(e.Logger),
			)
			if err != nil {
				return err
			}

			if upToDate && preCondMet {
				if e.Verbose || (!call.Silent && !t.Silent && !e.Taskfile.Silent && !e.Silent) {
					e.Logger.Errf(logger.Magenta, "task: Task %q is up to date\n", t.Name())
				}
				return nil
			}
		}

		for _, p := range t.Prompt {
			if p != "" && !e.Dry {
				if err := e.Logger.Prompt(logger.Yellow, p, "n", "y", "yes"); errors.Is(err, logger.ErrNoTerminal) {
					return &errors.TaskCancelledNoTerminalError{TaskName: call.Task}
				} else if errors.Is(err, logger.ErrPromptCancelled) {
					return &errors.TaskCancelledByUserError{TaskName: call.Task}
				} else if err != nil {
					return err
				}
			}
		}

		if err := e.mkdir(t); err != nil {
			e.Logger.Errf(logger.Red, "task: cannot make directory %q: %v\n", t.Dir, err)
		}

		var deferredExitCode uint8

		if t.Cmds[cmdIndex].Defer {
			e.runDeferred(t, call, cmdIndex, &deferredExitCode)
		} else {
			if err := e.runCommand(ctx, t, call, cmdIndex); err != nil {
				if err2 := e.statusOnError(t); err2 != nil {
					e.Logger.VerboseErrf(logger.Yellow, "task: error cleaning status on error: %v\n", err2)
				}

				exitCode, isExitError := interp.IsExitStatus(err)
				if isExitError {
					if t.IgnoreError {
						e.Logger.VerboseErrf(logger.Yellow, "task: task error ignored: %v\n", err)
						return nil
					}
					deferredExitCode = exitCode
				}

				if call.Indirect {
					return err
				}

				return &errors.TaskRunError{TaskName: t.Task, Err: err}
			}
			e.Logger.VerboseErrf(logger.Magenta, "task: %q finished\n", call.Task)
		}
		return nil
	})
}
