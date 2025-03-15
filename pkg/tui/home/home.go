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

package home

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hypershift-community/hyper-console/pkg/tui/lib/keys"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/simplelist"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/styles"
)

const (
	defaultWidth  = 30
	defaultHeight = 30
)

type SelectMessage struct {
	Selected int
}

type Model struct {
	list   list.Model
	keyMap *keys.KeyMap
}

func New() tea.Model {
	items := []simplelist.Item{
		{Name: "Recipes", Description: "View and run recipes"},
		{Name: "HyperShift Clusters", Description: "View and manage HyperShift clusters"},
	}

	defaultStyles := styles.DefaultStyles()
	keyMap := keys.NewListKeyMap()

	l := simplelist.NewList(keyMap, &defaultStyles, defaultWidth, defaultHeight, items...)

	l.Title = "HyperShift Dev Console"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = defaultStyles.Pagination
	l.Styles.HelpStyle = defaultStyles.Help

	return &Model{
		list:   l,
		keyMap: keyMap,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
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
		// We only handle the enter key cause the list will handel the
		// navigation keys when we call m.list.Update(msg)
		case m.keyMap.Matches(msg, keys.Enter):
			cmd = selectCmd(m.list.Cursor())
		}
		cmds = append(cmds, cmd)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return "\n" + m.list.View()
}

func selectCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return SelectMessage{Selected: index}
	}
}
