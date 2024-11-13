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

package keys

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type KeyType int

const (
	CursorUp KeyType = iota
	CursorDown
	Enter
	Create
	Delete
	Cancel
	Quit
	ForceQuit
)

func (k KeyType) String() string {
	return [...]string{"CursorUp", "CursorDown", "Enter", "Create", "Delete", "Cancel", "Quit", "ForceQuit"}[k]
}

func (k KeyType) KeyName() string {
	return [...]string{"ctrl+k", "ctrl+j", "enter", "ctrl+a", "ctrl+d", "esc", "q", "ctrl+c"}[k]
}

func (k KeyType) Help() string {
	return [...]string{
		"Move up",
		"Move down",
		"Select the currently highlighted item",
		"Create a new item",
		"Delete the currently selected item",
		"Cancel",
		"Quit",
		"Force quit",
	}[k]
}

//type KeyMap struct {
//	CursorUp   key.Binding
//	CursorDown key.Binding
//	Enter      key.Binding
//	Create     key.Binding
//	Delete     key.Binding
//	Cancel     key.Binding
//	Quit       key.Binding
//	ForceQuit  key.Binding
//
//	State string
//}

type KeyMap struct {
	items     map[KeyType]key.Binding
	shortHelp []key.Binding
}

func NewKeyMap() *KeyMap {
	return &KeyMap{
		items: make(map[KeyType]key.Binding),
	}
}

func NewDefaultKeyMap() *KeyMap {
	return NewKeyMap().
		Add(CursorUp, CursorUp.Help()).
		Add(CursorDown, CursorDown.Help()).
		Add(Enter, Enter.Help()).
		Add(Create, Create.Help()).
		Add(Delete, Delete.Help()).
		Add(Cancel, Cancel.Help()).
		Add(Quit, Quit.Help()).
		Add(ForceQuit, ForceQuit.Help()).
		WithShortHelp(CursorUp, CursorDown)

	//items := map[KeyType]key.Binding{
	//	CursorUp: key.NewBinding(
	//		key.WithKeys("ctrl+k"),
	//		key.WithHelp("ctrl+k", "move up"),
	//	),
	//	CursorDown: key.NewBinding(
	//		key.WithKeys("ctrl+j"),
	//		key.WithHelp("ctrl+j", "move down"),
	//	),
	//	Enter: key.NewBinding(
	//		key.WithKeys("enter"),
	//		key.WithHelp("enter", "Check out the currently selected branch"),
	//	),
	//	Create: key.NewBinding(
	//		key.WithKeys("ctrl+a"),
	//		key.WithHelp(
	//			"ctrl+a",
	//			"Create a new branch, with confirmation",
	//		),
	//	),
	//	Delete: key.NewBinding(
	//		key.WithKeys("ctrl+d"),
	//		key.WithHelp(
	//			"ctrl+d",
	//			"Delete the currently selected branch, with confirmation",
	//		),
	//	),
	//
	//	Cancel: key.NewBinding(
	//		key.WithKeys("esc"),
	//		key.WithHelp("esc", "Cancel"),
	//	),
	//}
	//return &KeyMap{
	//	items: items,
	//	shortHelp: []key.Binding{
	//		items[CursorUp],
	//		items[CursorDown],
	//	},
	//}
}

func (k *KeyMap) ShortHelp() []key.Binding {
	return k.shortHelp
}

func (k *KeyMap) FullHelp() [][]key.Binding {
	// TODO: Fixme. Why is this returning a slice of slices?
	inner := make([]key.Binding, len(k.items))
	i := 0
	for _, v := range k.items {
		inner[i] = v
		i++
	}
	return [][]key.Binding{inner}
}

//func (m *Model) updateKeybindins() {
//
//	switch m.currentUI {
//	case HomeUI:
//		m.keyMap.Enter.SetEnabled(true)
//		m.keyMap.Create.SetEnabled(true)
//		m.keyMap.Delete.SetEnabled(true)
//
//		m.keyMap.Cancel.SetEnabled(false)
//
//	default:
//		m.keyMap.Enter.SetEnabled(true)
//		m.keyMap.Create.SetEnabled(true)
//		m.keyMap.Delete.SetEnabled(true)
//		m.keyMap.Cancel.SetEnabled(false)
//	}
//}

func (k *KeyMap) Matches(msg tea.KeyMsg, keyType KeyType) bool {
	return key.Matches(msg, k.items[keyType])
}

func (k *KeyMap) Add(keyType KeyType, help string) *KeyMap {
	k.items[keyType] = key.NewBinding(
		key.WithKeys(keyType.KeyName()),
		key.WithHelp(keyType.KeyName(), help),
	)
	return k
}

func (k *KeyMap) WithShortHelp(keyType ...KeyType) *KeyMap {
	k.shortHelp = make([]key.Binding, len(keyType))
	for i, kt := range keyType {
		k.shortHelp[i] = k.items[kt]
	}
	return k
}
