//go:build mage

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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	moduleName = "github.com/hypershift-community/hyper-console"
	binaryName = "bin/hyperdev"

	// Path to the go-task/task module.
	// This should point to our fork of the task module.
	taskModulePath = "../task"

	// Destination path to copy the task module to.
	embededTaskPackage = "pkg/task"

	// When embedding the task module, we'll replace the following import paths.
	taskFileOriginalImport = "github.com/go-task/task/v3"
)

var (
	// We'll replace the above with this import path.
	// Note: When patching the fork, we'll revert this back to the original.
	embeddedTaskPackageImport = filepath.Join(moduleName, embededTaskPackage)

	// Files or directories that should NOT be copied over from the task module.
	excludedPaths = []string{
		".git",
		".github",
		".gitignore",
		".gitattributes",
		".vscode",
		".idea",
		".editorconfig",
		".golangci.yml",
		".goreleaser.yml",
		".mockery.yaml",
		".nvmrc",
		".prettierrc.yml",
		"CHANGELOG.md",
		"install-task.sh",
		"package.json",
		"package-lock.json",
		"README.md",
		"Taskfile.yml",
		"website",
		"bin",
		"testdata",
		"*_test.go",
		"go.mod",
		"go.sum",
	}

	goexec = mg.GoCmd()
	g0     = sh.RunCmd(goexec)
)

var Default = Run

// Build builds the binary
func Build() error {
	fmt.Println("Building...")
	return g0("build", "-o", binaryName, "./cmd/main.go")
}

// Clean cleans up the built binary
func Clean() {
	fmt.Println("Cleaning...")
	_ = os.RemoveAll(binaryName)
}

// Run runs the code without building a binary first
func Run() {
	cmdArgs := append([]string{"run", "cmd/main.go"})
	if len(os.Args) > 2 {
		cmdArgs = append(cmdArgs, os.Args[2:]...)
	}
	// We want to always print the output of the command to stdout
	_ = sh.RunV(goexec, cmdArgs...)
	return
}

// Embedfork copies our local fork of go-task/task from taskModulePath into embededTaskPackage,
// then renames `internal` -> `lib`, and rewrites imports to reference our
// embedded paths rather than the upstream.
func Embedfork() error {
	mg.Deps(ensureEmbeddedDir)

	// Check if the embedded task has local modifications
	hasChanges, out, err := hasLocalChanges("", embededTaskPackage)
	if err != nil {
		return fmt.Errorf("failed to embed fork: %w", err)
	}
	// If there's any output, it means local modifications exist
	if hasChanges {
		return fmt.Errorf("The embedded task package at %s has local modifications.\n git status:\n%s\n(Note: You can sync back to the fork using mage patchfork)", embededTaskPackage, out)
	}

	rsyncArgs := []string{"-av", "--delete"}
	for _, e := range excludedPaths {
		rsyncArgs = append(rsyncArgs, "--exclude", e)
	}
	rsyncArgs = append(rsyncArgs, taskModulePath+"/", embededTaskPackage+"/")

	fmt.Printf("Syncing from %s to %s...\n", taskModulePath, embededTaskPackage)
	if err := sh.RunV("rsync", rsyncArgs...); err != nil {

		return fmt.Errorf("rsync failed: %w", err)
	}

	// Rewrite import paths in .go files
	fmt.Println("Rewriting imports to embedded paths...")

	if err := rewriteImports(embededTaskPackage, taskFileOriginalImport, embeddedTaskPackageImport); err != nil {
		return err
	}

	fmt.Println("Embedfork complete.")
	return nil
}

// Patchfork copies local changes from pkg/task to ../task,
// then reverts your local "myorg/myrepo/pkg/task" imports back to "github.com/go-task/task/v3" in the fork,
// and finally stashes changes in pkg/task.
// This target will fail if the fork has local modifications. Use forcepatch to force the patching.
func Patchfork() error {

	// Check if the fork has local modifications
	hasChanges, out, err := hasLocalChanges(taskModulePath, "")
	if err != nil {
		return fmt.Errorf("failed to patch fork: %w", err)
	}
	// If there's any output, it means local modifications exist
	if hasChanges {
		return fmt.Errorf("The fork at %s has local modifications.\n git status:\n%s\nUse forcepatch to force the patching.", taskModulePath, out)
	}

	return Forcepatch()
}

// Forcepatch copies local changes from pkg/task to ../task while overwriting any local changes in the fork,
// then reverts your local "myorg/myrepo/pkg/task" imports back to "github.com/go-task/task/v3" in the fork,
// and finally stashes changes in pkg/task.
func Forcepatch() error {

	stash, _, err := hasLocalChanges("", embededTaskPackage)
	if err != nil {
		return fmt.Errorf("failed to forcepatch: %w", err)
	}

	rsyncArgs := []string{"-av", "--delete"}
	for _, e := range excludedPaths {
		rsyncArgs = append(rsyncArgs, "--exclude", e)
	}
	rsyncArgs = append(rsyncArgs, embededTaskPackage+"/", taskModulePath+"/")

	fmt.Printf("Syncing all changes from %s to %s...\n", embededTaskPackage, taskModulePath)
	if err := sh.RunV("rsync", rsyncArgs...); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}
	// Rewrite import paths in the fork back to the "old" module path
	// so that the fork remains consistent with its original references.
	fmt.Println("Rewriting imports in the fork back to original module paths...")
	if err := rewriteImports(taskModulePath, embeddedTaskPackageImport, taskFileOriginalImport); err != nil {
		return err
	}

	if stash {
		// Stash local changes under embededTaskPackage.
		fmt.Printf("Stashing local modifications in %s\n...", embededTaskPackage)
		stashMsg := fmt.Sprintf("Stash %s changes after patching the fork", embededTaskPackage)
		if err := sh.RunV("git", "stash", "push", "--include-untracked",
			"-m", stashMsg, "--", embededTaskPackage); err != nil {
			return fmt.Errorf("failed to stash %s changes: %w", embededTaskPackage, err)
		}
		fmt.Println("\nStash content:\n")
		sh.RunV("git", "stash", "show")
		sh.RunV("git", "stash", "show", "-p")
	}
	fmt.Println("Patching the fork complete.")
	return nil
}

// rewriteImports uses sed to replace oldImport with newImport in all .go files under baseDir
// keeping in mind that the sed flags differ between macOS and Linux.
func rewriteImports(baseDir, oldImport, newImport string) error {
	fmt.Printf("Replacing %q → %q in Go source files...\n", oldImport, newImport)

	// Example cross-platform usage:
	//   - On Linux: sed -i 's/old/new/g' file
	//   - On macOS: sed -i '' 's/old/new/g' file
	var sedArg string
	if runtime.GOOS == "darwin" {
		// macOS
		sedArg = "-i ''"
	} else {
		// Linux, others
		sedArg = "-i"
	}

	// We wanna run something like:
	//   find pkg/task -type f -name '*.go' -exec sed -i '' 's|github.com/go-task/task/v3|github.com/example/project/pkg/embedded|g' {} +
	cmd := fmt.Sprintf(
		`find %s -type f -name '*.go' -exec sed %s 's|%s|%s|g' {} +`,
		baseDir, sedArg, oldImport, newImport,
	)

	if err := sh.Run("bash", "-c", cmd); err != nil {
		return fmt.Errorf("failed to rewrite imports for %q: %w", oldImport, err)
	}
	tea.Sequence()
	return nil
}

// ensureEmbeddedDir creates the destination directory if it doesn't already exist.
func ensureEmbeddedDir() error {
	if err := os.MkdirAll(embededTaskPackage, 0o755); err != nil {
		return fmt.Errorf("creating destination dir %q failed: %w", embededTaskPackage, err)
	}
	return nil
}

func hasLocalChanges(repoPath, localPath string) (bool, string, error) {

	if repoPath == "" {
		repoPath = "."
	}
	gitArgs := []string{"-C", repoPath, "status", "--porcelain"}

	if localPath != "" {
		gitArgs = append(gitArgs, "--", localPath)
	}

	out, err := sh.Output("git", gitArgs...)
	if err != nil {
		return false, "", fmt.Errorf("failed to check git status: %w", err)
	}
	return strings.TrimSpace(out) != "", out, nil
}
