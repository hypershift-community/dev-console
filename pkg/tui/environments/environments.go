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

package environments

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hypershift-community/hyper-console/pkg/config"
	"github.com/hypershift-community/hyper-console/pkg/env"
	"github.com/hypershift-community/hyper-console/pkg/logging"
	"github.com/hypershift-community/hyper-console/pkg/recipes"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/keys"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/navigation"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/simplelist"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/styles"
)

var Logger = logging.Logger

type SelectMessage struct {
	Environment string
}

type envsLoadedMessage map[string]*env.Env

type Model struct {
	list        list.Model
	recipe      *recipes.Recipe
	cfg         *config.Config
	envs        map[string]*env.Env
	keyMap      *keys.KeyMap
	initialized bool
	err         error
}

func New(windowWidth int, windowHeight int, recipe *recipes.Recipe, cfg *config.Config) tea.Model {
	defaultStyles := styles.DefaultStyles()
	keyMap := keys.NewListKeyMap()

	l := simplelist.NewList(keyMap, &defaultStyles, windowWidth, windowHeight)

	l.Title = fmt.Sprintf("Set environment for recipe: %s", recipe.Name)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = defaultStyles.Pagination
	l.Styles.HelpStyle = defaultStyles.Help

	return &Model{
		list:   l,
		recipe: recipe,
		cfg:    cfg,
		keyMap: keyMap,
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		Logger.Debug("Loading environments")
		envs, err := env.LoadAll(m.cfg.EnvironmentsDir)
		if err != nil {
			m.err = err
			return nil
		}
		return envsLoadedMessage(envs)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch {
		case m.keyMap.Matches(msg, keys.Enter):
			cmd = m.getSelectedCmd()
		case m.keyMap.Matches(msg, keys.Cancel):
			return m, navigation.Back()
		}
		cmds = append(cmds, cmd)
	case envsLoadedMessage:
		envs := msg
		items := make([]list.Item, len(envs))
		i := 0
		for n, e := range envs {
			//TODO: this means we know that simplelist.Item is a list.Item. This is not ideal and should be fixed.
			item := simplelist.Item{Name: n}
			if e.Description != "" {
				item.Description = e.Description
			}
			items[i] = &item
			i++
		}
		m.list.SetItems(items)
		m.envs = envs
		m.initialized = true
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if len(m.envs) == 0 {
		if m.initialized {
			return "\nNo environments found in " + m.cfg.EnvironmentsDir
		} else {
			if m.err != nil {
				return "\nError loading environments: " + m.err.Error()
			}
			return "\nLoading environments..."
		}
	}

	return "\n" + m.list.View()
}

func (m *Model) getSelectedCmd() tea.Cmd {
	return func() tea.Msg {
		items := m.list.Items()
		if len(items) == 0 {
			Logger.Debug("No environments to select from")
			return nil
		}
		Logger.Debug("Environment selected.", "environment", items[m.list.Cursor()].(*simplelist.Item).Name)
		envName := items[m.list.Cursor()].(*simplelist.Item).Name
		return SelectMessage{Environment: envName}
	}
}
