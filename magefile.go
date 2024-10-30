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

//go:build mage

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "hyperdev"
)

var (
	goexec = mg.GoCmd()
	g0     = sh.RunCmd(goexec)
)

func mustRun(cmd string, args ...string) {
	out := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("\n> %s %s\n", cmd, strings.Join(args, " ")),
	)

	fmt.Println(out)
	if err := sh.RunV(cmd, args...); err != nil {
		panic(err)
	}
}

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
