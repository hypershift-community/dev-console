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
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"hypershift-dev-console/pkg/env"
)

func Test_task_Execute(t *testing.T) {
	tests := []struct {
		name          string
		taskYaml      string
		outValidator  func(t *testing.T, out string)
		expectedError string
		env           *env.Env
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
			outValidator: func(t *testing.T, out string) {
				require.Contains(t, out, "bar")
			},
		}, {
			name: "setting an env var should override the default value",
			taskYaml: `version: '3'
env:
  FOO: bar
tasks:
  default:
    cmds:
      - echo $FOO
`,
			outValidator: func(t *testing.T, out string) {
				require.Equal(t, "baz\n", out)
			},
			env: &env.Env{
				Vars: map[string]string{
					"FOO": "baz",
				},
			},
		},

		{
			name:          "running an invalid task should fail",
			taskYaml:      `version: '3' env: FOO: bar tasks:`,
			expectedError: "error setting up task executor: task: Failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)

			// write the task file
			err := writeTaskFile(dir, tt.taskYaml)
			require.NoError(t, err)

			// create the task
			task := NewTask(dir, nil, stdout, stderr)

			if tt.env != nil {
				task.SetEnv(tt.env)
			}

			// execute the task
			err = task.Execute()
			if tt.expectedError != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.outValidator != nil {
					tt.outValidator(t, stdout.String())
				}
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
