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

package taskexec

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/hypershift-community/hyper-console/pkg/env"
	"github.com/hypershift-community/hyper-console/pkg/iter"
	"github.com/hypershift-community/hyper-console/pkg/task"
	"github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"
)

type ExecutorIterator interface {
	iter.Iterable[Executor]
	GetTask() *ast.Task
}

type Executor interface {
	Execute() error
	SetEnv(env *env.Env)
	SetIO(stdin io.Reader, stdout, stderr io.Writer)
}

// TaskOption is a function that configures a task executor.
type TaskOption func(Executor)

// WithEnv sets the environment variables for the task executor.
func WithEnv(env *env.Env) TaskOption {
	return func(e Executor) {
		e.SetEnv(env)
	}
}

func WithIO(stdin io.Reader, stdout, stderr io.Writer) TaskOption {
	return func(e Executor) {
		e.SetIO(stdin, stdout, stderr)
	}
}

type _task struct {
	task.Executor
	env      *env.Env
	cmdIndex int
	prepared bool
	task     *ast.Task
	call     *task.Call
}

func NewExecutorIterator(dir string, opts ...TaskOption) (ExecutorIterator, int, error) {
	t := &_task{
		Executor: task.Executor{
			Dir:    dir,
			Stdin:  &bytes.Buffer{},
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		},
	}
	for _, opt := range opts {
		opt(t)
	}
	if !t.prepared {
		// Set up the executor
		if err := t.Setup(); err != nil {
			return nil, -1, fmt.Errorf("error setting up task executor: %w", err)
		}

		if t.env != nil {
			for k, v := range t.env.Vars {
				taskEnv := t.Compiler.TaskfileEnv
				if _, ok := taskEnv.Get(k); ok {
					taskEnv.Set(k, ast.Var{Value: v})
				}
			}
		}
		// Define the _task to run
		call := &task.Call{Task: "default"}

		task, err := t.PrepareTask(call)
		if err != nil {
			return nil, -1, fmt.Errorf("error perapring execution of task: %w", err)
		}
		t.task = task
		t.call = call
		t.prepared = true
	}
	return t, len(t.task.Cmds), nil
}

func (t *_task) HasNext() bool {
	return t.cmdIndex < len(t.task.Cmds)
}

func (t *_task) Next() (Executor, error) {

	if t.cmdIndex >= len(t.task.Cmds) {
		return nil, fmt.Errorf("no more commands to run")
	}
	t.cmdIndex++
	return t, nil
}

func (t *_task) Execute() error {
	if t.cmdIndex > len(t.task.Cmds) {
		return fmt.Errorf("no more commands to run")
	}
	// Run the _task
	//TODO: Weave in a context
	if err := t.RunTaskCmd(context.Background(), t.call, t.task, t.cmdIndex-1); err != nil {
		return fmt.Errorf("error running task: %w", err)
	}
	return nil
}

func (t *_task) SetEnv(env *env.Env) {
	t.env = env
}

func (t *_task) SetIO(stdin io.Reader, stdout, stderr io.Writer) {
	t.Stdin = stdin
	t.Stdout = stdout
	t.Stderr = stderr
}

func (t *_task) GetTask() *ast.Task {
	return t.task
}
