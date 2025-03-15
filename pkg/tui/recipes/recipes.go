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
	"github.com/charmbracelet/lipgloss"

	"github.com/hypershift-community/hyper-console/pkg/config"
	"github.com/hypershift-community/hyper-console/pkg/recipes"
	"github.com/hypershift-community/hyper-console/pkg/tui/environments"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/keys"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/navigation"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/styles"
)

var (
	SetEnvKey = keys.NewCustomKey("Set Environment", "ctrl+e", "Set the environment for the recipe")
)

type SelectMessage struct {
	Recipe recipes.Recipe
}

type SetEnvMessage struct {
	Selected int
	Recipe   *recipes.Recipe
}

type recipesMessage []recipes.Recipe

type item struct {
	Name        string
	Description string
	Dir         string
	CurrentEnv  string
}

func (i *item) FilterValue() string { return i.Name }

type Model struct {
	list         list.Model
	keyMap       *keys.KeyMap
	recipes      []recipes.Recipe
	cfg          *config.Config
	delegate     *itemDelegate
	windowWidth  int
	windowHeight int
}

func New(width, height int, cfg *config.Config) tea.Model {
	items := make([]list.Item, 0)
	defaultStyles := styles.DefaultStyles()
	keyMap := keys.NewListKeyMap().
		WithKey(SetEnvKey, true)
	delegate := newItemDelegate(keyMap, &defaultStyles)
	l := list.New(items, delegate, width, height)
	l.Title = "HyperShift Dev Console"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = defaultStyles.Pagination
	l.Styles.HelpStyle = defaultStyles.Help

	return &Model{
		list:         l,
		keyMap:       keyMap,
		recipes:      make([]recipes.Recipe, 0),
		cfg:          cfg,
		delegate:     delegate,
		windowWidth:  width,
		windowHeight: height,
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
		case m.keyMap.Matches(msg, keys.Enter):
			cmd = m.getSelectedCmd()
		case m.keyMap.Matches(msg, keys.Cancel):
			return m, navigation.Back()
		case m.keyMap.Matches(msg, SetEnvKey):
			return m, m.setEnvCmd(m.list.Cursor())
		}
		cmds = append(cmds, cmd)
	case recipesMessage:
		m.recipes = msg
		m.refreshList()
	case environments.SelectMessage:
		m.recipes[m.list.Cursor()].Environment = msg.Environment
		m.refreshList()
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return "\n" + m.list.View()
	//return lipgloss.PlaceHorizontal(m.windowWidth, lipgloss.Center, listView)
}

func (m *Model) getSelectedCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectMessage{Recipe: m.recipes[m.list.Cursor()]}
	}
}

func (m *Model) setEnvCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return SetEnvMessage{Recipe: &m.recipes[index]}
	}
}

func (m *Model) refreshList() {
	items := make([]list.Item, len(m.recipes))
	widest := 0
	for i, r := range m.recipes {
		it := &item{
			Name:        r.Name,
			Description: r.Description,
			Dir:         r.Dir,
			CurrentEnv:  r.Environment,
		}
		renderedItem := m.delegate.renderItemDisplay(it, i == m.list.Cursor())
		w := lipgloss.Width(renderedItem)
		if w > widest {
			widest = w
		}
		items[i] = it
	}
	m.list.SetWidth(widest)
	m.list.SetItems(items)
}
