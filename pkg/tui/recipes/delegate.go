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
	i, ok := listItem.(*item)
	if !ok {
		return
	}
	renderedItem := d.renderItemDisplay(i, m.Index() == index)
	fmt.Fprint(w, renderedItem)
}

func (d *itemDelegate) ShortHelp() []key.Binding {
	return d.keys.ShortHelp()
}

func (d *itemDelegate) FullHelp() [][]key.Binding {
	return d.keys.FullHelp()
}

func (d *itemDelegate) renderItemDisplay(i *item, selected bool) string {
	var name string
	var desc string
	var prefix string
	var suffix string
	nameStyle := d.styles.NormalTitle

	if selected {
		prefix = "> "
		nameStyle = d.styles.SelectedTitle
	}

	if i.CurrentEnv != "" {
		suffix = fmt.Sprintf("%s [Env: %s]", i.Name, i.CurrentEnv)
	}

	name = nameStyle.Render(fmt.Sprintf("%s%s%s", prefix, i.Name, suffix))
	desc = d.styles.NormalDesc.Render(i.Description)

	return fmt.Sprintf("%s - %s", name, desc)
}
