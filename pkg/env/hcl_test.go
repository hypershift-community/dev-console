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

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hypershift-community/hyper-console/pkg/task/errors"
)

func TestLoad(t *testing.T) {
	for _, tt := range []struct {
		name    string
		EnvFile string
		EnvDir  map[string]string
		WantEnv map[string]string
		Err     error
	}{
		{
			name: "Load environment with a simple valid hcl file",
			EnvFile: `
NAMESPACE = "clusters"
CLUSTER_NAME = "cluster1"
REPLICAS = 3
`,
			EnvDir: nil,
			WantEnv: map[string]string{
				"NAMESPACE":    "clusters",
				"CLUSTER_NAME": "cluster1",
				"REPLICAS":     "3",
			},
		},
		{
			name: "Load environment with a valid hcl file and a file in the env.d directory",
			EnvFile: `
NAMESPACE = "clusters"
CLUSTER_NAME = "cluster1"
REPLICAS = 3
`,
			EnvDir: map[string]string{
				"KUBECONFIG": "THIS IS A VALID KUBECONFIG FILE",
			},
			WantEnv: map[string]string{
				"NAMESPACE":    "clusters",
				"CLUSTER_NAME": "cluster1",
				"REPLICAS":     "3",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory
			dir := t.TempDir()

			// Create the env file
			envFile := filepath.Join(dir, DefaultEnvFile)
			if err := createFile(envFile, tt.EnvFile); err != nil {
				t.Fatalf("failed to create env file: %v", err)
			}

			if tt.EnvDir != nil {
				// Create the env.d directory
				envDir := filepath.Join(dir, DefaultEnvDir)
				if err := createDir(envDir); err != nil {
					t.Fatalf("failed to create env.d directory: %v", err)
				}

				// Create the env.d files
				for k, v := range tt.EnvDir {
					file := filepath.Join(envDir, k)
					if err := createFile(file, v); err != nil {
						t.Fatalf("failed to create env.d file: %v", err)
					}
				}
			}

			// Load the environment
			env, err := Load(dir)
			if tt.Err != nil {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.Err, err)
				} else if !errors.Is(err, tt.Err) {
					t.Fatalf("expected error %v, got %v", tt.Err, err)
				}
			}

			if env.Name != filepath.Base(dir) {
				t.Fatalf("expected env name %s, got %s", filepath.Base(dir), env.Name)
			}

			// Compare the environment
			for k, v := range tt.WantEnv {
				if env.Vars[k] != v {
					t.Fatalf("expected env var %s to be %s, got %s", k, v, env.Vars[k])
				}
			}
			if tt.EnvDir != nil {
				for k := range tt.EnvDir {
					if _, ok := env.Vars[k]; !ok {
						t.Fatalf("expected env var %s to be present", k)
					}
					file := filepath.Join(dir, DefaultEnvDir, k)
					if env.Vars[k] != file {
						t.Fatalf("expected env var %s to be %s, got %s", k, file, env.Vars[k])
					}
				}
			}
			fmt.Printf("env: %v\n", env)
		})
	}
}

func createFile(file, content string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func createDir(dir string) error {
	return os.Mkdir(dir, 0755)
}
