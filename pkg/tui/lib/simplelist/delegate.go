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

package simplelist

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hypershift-community/hyper-console/pkg/tui/lib/keys"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/styles"
)

type itemDelegate struct {
	keys   *keys.KeyMap
	styles *styles.Styles
}

func newItemDelegate(keys *keys.KeyMap, styles *styles.Styles) *itemDelegate {
	return &itemDelegate{
		keys:   keys,
		styles: styles,
	}
}

func (d *itemDelegate) Height() int                               { return 1 }
func (d *itemDelegate) Spacing() int                              { return 0 }
func (d *itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*Item)
	if !ok {
		return
	}

	var name string
	var desc string

	if index == m.Index() {
		name = d.styles.SelectedTitle.Render("> " + i.Name)
		desc = d.styles.SelectedDesc.Render(i.Description)
	} else {
		name = d.styles.NormalTitle.Render(i.Name)
		desc = d.styles.NormalDesc.Render(i.Description)
	}
	var itemListStyle string
	if i.Description == "" {
		itemListStyle = name
	} else {
		itemListStyle = fmt.Sprintf("%s - %s", name, desc)
	}

	fmt.Fprint(w, itemListStyle)
}

func (d *itemDelegate) ShortHelp() []key.Binding {
	return d.keys.ShortHelp()
}

func (d *itemDelegate) FullHelp() [][]key.Binding {
	return d.keys.FullHelp()
}
