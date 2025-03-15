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

package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hypershift-community/hyper-console/pkg/config"
	"github.com/hypershift-community/hyper-console/pkg/tui/environments"
	"github.com/hypershift-community/hyper-console/pkg/tui/home"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/navigation"
	"github.com/hypershift-community/hyper-console/pkg/tui/recipes"
	"github.com/hypershift-community/hyper-console/pkg/tui/recipes/run"
)

type Model struct {
	modelStack []tea.Model
	windowSize tea.WindowSizeMsg
	cfg        *config.Config
}

func NewModel(cfg *config.Config) tea.Model {

	return &Model{
		modelStack: []tea.Model{home.New()},
		cfg:        cfg,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	var model tea.Model

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
	case home.SelectMessage:
		model = recipes.New(m.windowSize.Width, m.windowSize.Height, m.cfg)
		cmds = append(cmds, model.Init())
		m.modelStack = append(m.modelStack, model)
	case recipes.SelectMessage:
		model = run.New(m.windowSize.Width, m.windowSize.Height, msg.Recipe, m.cfg.EnvironmentsDir)
		cmds = append(cmds, model.Init())
		m.modelStack = append(m.modelStack, model)
	case recipes.SetEnvMessage:
		model = environments.New(m.windowSize.Width, m.windowSize.Height, msg.Recipe, m.cfg)
		cmds = append(cmds, model.Init())
		m.modelStack = append(m.modelStack, model)
	case environments.SelectMessage:
		// All we need to do is pop the current model off the stack
		// and rely on passing the selected environment message to the
		// previous model
		m.modelStack = m.modelStack[:len(m.modelStack)-1]
	case navigation.BackMessage:
		if len(m.modelStack) > 1 {
			m.modelStack = m.modelStack[:len(m.modelStack)-1]
		}
	}

	m.modelStack[len(m.modelStack)-1], cmd = m.modelStack[len(m.modelStack)-1].Update(msg)

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return m.modelStack[len(m.modelStack)-1].View()
}
