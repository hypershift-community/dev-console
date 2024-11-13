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
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"hypershift-dev-console/pkg/config"
	"hypershift-dev-console/pkg/recipes"
	"hypershift-dev-console/pkg/tui/keys"
	"hypershift-dev-console/pkg/tui/navigation"
	"hypershift-dev-console/pkg/tui/styles"
)

type SelectMessage struct {
	Selected int
	Recipe   recipes.Recipe
}

type recipesMessage []recipes.Recipe

type item struct {
	Name        string
	Description string
	Dir         string
}

func (i *item) FilterValue() string { return i.Name }

type Model struct {
	list    list.Model
	keyMap  *keys.KeyMap
	recipes []recipes.Recipe
	cfg     *config.Config
}

func New(width, height int, cfg *config.Config) tea.Model {
	items := make([]list.Item, 0)
	styles := styles.DefaultStyles()
	keys := keys.NewDefaultKeyMap()
	l := list.New(items, newItemDelegate(keys, &styles), width, height)
	l.Title = "HyperShift Dev Console"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return &Model{
		list:    l,
		keyMap:  keys,
		recipes: make([]recipes.Recipe, 0),
		cfg:     cfg,
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		rcps, err := recipes.GetRecipes(m.cfg.RecipesDir)
		if err != nil {
			return nil
		}
		return recipesMessage(rcps)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch {
		case m.keyMap.Matches(msg, keys.CursorUp):
			m.list.CursorUp()
		case m.keyMap.Matches(msg, keys.CursorDown):
			m.list.CursorDown()
		case m.keyMap.Matches(msg, keys.Enter):
			cmd = m.selectCmd(m.list.Cursor())
		case m.keyMap.Matches(msg, keys.Cancel):
			return m, navigation.Back()
		}
		cmds = append(cmds, cmd)
	case recipesMessage:
		items := make([]list.Item, 0)
		for _, recipe := range msg {
			items = append(items, &item{Name: recipe.Name, Description: recipe.Description, Dir: recipe.Dir})
		}
		m.list.SetItems(items)
		m.recipes = msg
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return "\n" + m.list.View()
}

func (m *Model) selectCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return SelectMessage{Selected: index, Recipe: m.recipes[index]}
	}
}
