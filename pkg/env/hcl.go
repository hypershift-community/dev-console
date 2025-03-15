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

	"github.com/hashicorp/hcl/v2/hclsimple"

	"github.com/hypershift-community/hyper-console/pkg/logging"
)

// Env represents an environment configuration. Each environment consists of a directory
// which name is the environment name and an environment file called env.hcl which contains
// the environment configuration in HCL format.
//
// Any key-value pair in the env.hcl that starts with _INFO_ will be considered as metadata
// and will not be added to the environment vars map. This is useful for adding metadata to
// the environment configuration such as description or any other information that is not
// an environment variable. Currently, the only supported metadata key is _INFO_DESCRIPTION.
//
// For cases where you need an env var that points to a specific file "e.g. KUBECONFIG"
// that is different for each environment, you can create an env.d directory in the environment
// directory and place the file there. The file name should be the same as the env var name
// and the content will automatically be set to the path of the file.
//
// Example:
// ─── environments
//     ├── dev
//     │   ├── env.hcl
//     │   └── env.d
//     │       └── KUBECONFIG
//     └── prod
//         ├── env.hcl
//         └── env.d
//             └── KUBECONFIG
//
// Here, the KUBECONFIG env var will be set to the path of the file in the env.d directory.
//
// The reason for choosing HCL is that it will allow us to create our custom DSL that can be used
// to define the environment configuration.
//
// Example:
//
// cluster_ready = HCP.Status.Ready

const (
	// DefaultEnvFile is the default file path for the environment configuration
	DefaultEnvFile = "env.hcl"
	// DefaultEnvDir is the directory where environment specific files are placed
	DefaultEnvDir = "env.d"
)

var Logger = logging.Logger

type Env struct {
	Name        string
	Description string
	Vars        map[string]string
}

// Load loads the environment configuration from the specified directory. The directory should contain
// // subdirectories where each subdirectory is an environment. Each environment should contain
// // an env.hcl file and an optional env.d directory. For each file in the env.d directory, an
// // entry will be added to the environment vars map where the key is the file name and the value
// // is the absolute path to the file. This is useful for cases where you need an env var that
// // points to a specific file that is different for each environment "e.g. KUBECONFIG".
func Load(path string) (*Env, error) {
	var env Env
	envFile := filepath.Join(path, DefaultEnvFile)
	Logger.Debug("Loading environment configuration", "envFile", envFile)
	err := hclsimple.DecodeFile(envFile, nil, &env.Vars)
	if err != nil {
		Logger.Error("Error loading environment configuration", "envFile", envFile, "error", err)
		return nil, fmt.Errorf("error loading environment configuration: %w", err)
	}
	env.Name = filepath.Base(path)
	loadEnvDir(path, &env.Vars)
	if desc, ok := env.Vars["_INFO_DESCRIPTION"]; ok {
		env.Description = desc
		delete(env.Vars, "_INFO_DESCRIPTION")
	}
	return &env, nil
}

// LoadAll loads all environments in the specified directory. The directory should contain
// subdirectories where each subdirectory is an environment. Each environment should contain
// an env.hcl file and an optional env.d directory. For each file in the env.d directory, an
// entry will be added to the environment vars map where the key is the file name and the value
// is the absolute path to the file. This is useful for cases where you need an env var that
// points to a specific file that is different for each environment "e.g. KUBECONFIG".
func LoadAll(path string) (map[string]*Env, error) {
	environments := make(map[string]*Env)
	files, err := os.ReadDir(path)
	if err != nil {
		Logger.Error("Error reading environments directory", "path", path, "error", err)
		return nil, fmt.Errorf("error reading environments directory: %w", err)
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		env, err := Load(filepath.Join(path, file.Name()))
		if err != nil {
			Logger.Error("Error loading environment", "env", file.Name(), "error", err)
			return nil, fmt.Errorf("error loading environment %s: %w", file.Name(), err)
		}
		environments[file.Name()] = env
	}
	return environments, nil
}

func loadEnvDir(path string, vars *map[string]string) {
	envDir := filepath.Join(path, DefaultEnvDir)
	Logger.Debug("Loading environment directory", "envDir", envDir)
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		Logger.Debug("Environment directory does not exist", "envDir", envDir)
		return
	}
	files, err := os.ReadDir(envDir)
	if err != nil {
		Logger.Error("Error reading env.d directory", "envDir", envDir, "error", err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			Logger.Warn("Found directory in env.d directory. This is not supported.", "dir", file.Name())
			continue
		}
		path, err := filepath.Abs(filepath.Join(envDir, file.Name()))
		if err != nil {
			Logger.Error("Error getting absolute path for file", "file", file.Name(), "error", err)
			continue
		}
		(*vars)[file.Name()] = path
	}
}
