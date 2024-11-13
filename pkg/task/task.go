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
	"fmt"
	"io"

	taskfile "github.com/go-task/task/v3"
	"github.com/go-task/task/v3/taskfile/ast"
	"hypershift-dev-console/pkg/env"
)

type Executor interface {
	Execute() error
	SetEnv(env *env.Env)
}

type task struct {
	taskfile.Executor
	env *env.Env
}

func NewTask(dir string, stdin, stdout, stderr io.ReadWriter) Executor {
	return &task{
		Executor: taskfile.Executor{
			Dir:    dir,
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
	}
}

func (t *task) Execute() error {
	//TODO: Implement
	// Initialize the executor

	// Set up the executor
	if err := t.Setup(); err != nil {
		return fmt.Errorf("error setting up task executor: %w", err)
	}

	if t.env != nil {
		for k, v := range t.env.Vars {
			taskEnv := t.Executor.Compiler.TaskfileEnv
			if taskEnv.Exists(k) {
				taskEnv.Set(k, ast.Var{Value: v})
			}
		}
	}

	// Define the task to run
	call := &ast.Call{Task: "default"}
	// Run the task
	if err := t.RunTask(context.Background(), call); err != nil {
		return fmt.Errorf("error running task: %w", err)
	}
	return nil
}

func (t *task) SetEnv(env *env.Env) {
	t.env = env
}
