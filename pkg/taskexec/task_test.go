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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hypershift-community/hyper-console/pkg/env"
)

type validator func(t *testing.T, out string)

func Test_task_Execute(t *testing.T) {
	tests := []struct {
		name              string
		taskYaml          string
		cmdsCount         int
		validators        []validator
		expectedPrepError string
		env               *env.Env
	}{
		{
			name: "running a valid task with a single command should succeed",
			taskYaml: `version: '3'
env:
  FOO: bar
tasks:
  default:
    cmds:
      - echo $FOO
`,
			cmdsCount: 1,
			validators: []validator{func(t *testing.T, out string) {
				require.Contains(t, out, "bar")
			}},
		},
		{
			name: "running a valid task with multiple commands should succeed",
			taskYaml: `version: '3'
env:
  FOO: bar
tasks:
  default:
    cmds:
      - echo $FOO
      - echo second $FOO
`,
			cmdsCount: 2,
			validators: []validator{func(t *testing.T, out string) {
				require.Equal(t, out, "bar\n")
			}, func(t *testing.T, out string) {
				require.Equal(t, out, "second bar\n")
			}},
		},
		{
			name: "setting an env var should override the default value",
			taskYaml: `version: '3'
env:
  FOO: bar
  QUX: qux
tasks:
  default:
    cmds:
      - echo $FOO
      - echo second $QUX
`,
			cmdsCount: 2,
			validators: []validator{func(t *testing.T, out string) {
				require.Equal(t, "baz\n", out)
			}, func(t *testing.T, out string) {
				require.Equal(t, out, "second corge\n")
			}},
			env: &env.Env{
				Vars: map[string]string{
					"FOO": "baz",
					"QUX": "corge",
				},
			},
		},
		{
			name:              "running an invalid task should fail",
			taskYaml:          `version: '3' env: FOO: bar tasks:`,
			expectedPrepError: "error setting up task executor: task: Failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			// write the _task file
			err := writeTaskFile(dir, tt.taskYaml)
			require.NoError(t, err)

			var opts []TaskOption
			if tt.env != nil {
				opts = append(opts, WithEnv(tt.env))
			}
			taskIter, n, err := NewExecutorIterator(dir, opts...)
			if tt.expectedPrepError != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedPrepError)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.cmdsCount, n)
			for i := 0; i < tt.cmdsCount && taskIter.HasNext(); i++ {
				task, err := taskIter.Next()
				require.NoError(t, err)

				var stdout bytes.Buffer
				var stderr bytes.Buffer
				task.SetIO(nil, &stdout, &stderr)

				err = task.Execute()
				require.NoError(t, err)
				tt.validators[i](t, stdout.String())
			}
		})
	}
}

func writeTaskFile(dir, content string) error {
	fileName := fmt.Sprintf("%s/Taskfile.yml", dir)
	fd, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString(content)
	return err
}
