package taskfile

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/hypershift-community/hyper-console/pkg/task/internal/filepathext"
	"github.com/hypershift-community/hyper-console/pkg/task/internal/templater"
	"github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"
)

func Dotenv(vars *ast.Vars, tf *ast.Taskfile, dir string) (*ast.Vars, error) {
	env := ast.NewVars()
	cache := &templater.Cache{Vars: vars}

	for _, dotEnvPath := range tf.Dotenv {
		dotEnvPath = templater.Replace(dotEnvPath, cache)
		if dotEnvPath == "" {
			continue
		}
		dotEnvPath = filepathext.SmartJoin(dir, dotEnvPath)

		if _, err := os.Stat(dotEnvPath); os.IsNotExist(err) {
			continue
		}

		envs, err := godotenv.Read(dotEnvPath)
		if err != nil {
			return nil, fmt.Errorf("error reading env file %s: %w", dotEnvPath, err)
		}
		for key, value := range envs {
			if _, ok := env.Get(key); !ok {
				env.Set(key, ast.Var{Value: value})
			}
		}
	}

	return env, nil
}
