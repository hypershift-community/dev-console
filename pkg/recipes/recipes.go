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

package recipes

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RecipeInfo struct {
	Version     string `yaml:"version"`
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display-name"`
	Description string `yaml:"description"`
	Environment string `yaml:"environment,omitempty"`
}

type Recipe struct {
	RecipeInfo
	Dir string
}

func GetRecipes(recipesDir string) ([]Recipe, error) {
	var recipes []Recipe

	err := filepath.Walk(recipesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error reading recipes directory: %w", err)
		}

		if info.IsDir() {
			infoFilePath := filepath.Join(path, "info.yaml")
			if _, err := os.Stat(infoFilePath); err == nil {
				data, err := os.ReadFile(infoFilePath)
				if err != nil {
					return fmt.Errorf("error reading recipe info %s file: %w", infoFilePath, err)
				}

				var recipeInfo RecipeInfo
				if err := yaml.Unmarshal(data, &recipeInfo); err != nil {
					return fmt.Errorf("error unmarshalling recipe info %s file: %w", infoFilePath, err)
				}

				recipes = append(recipes, Recipe{recipeInfo, path})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return recipes, nil
}

func (r *Recipe) Run() {
	fmt.Printf("Running recipe %s\n", r.Name)
}
