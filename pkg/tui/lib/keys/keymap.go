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
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type defaultKey int

const (
	Up defaultKey = iota
	Down
	PageUp
	PageDown
	HalfPageUp
	HalfPageDown
	Enter
	Create
	Delete
	Cancel
	Quit
	ForceQuit
	//Unknown must stay as the last key. Add new default keys above it.
	Unknown
)

type KeyProvider interface {
	fmt.Stringer
	KeyStroke() string
	Help() string
	ShortHelp() string
}

func (k defaultKey) validate() {
	if k >= Unknown {
		panic("Unknown key")
	}
}

func (k defaultKey) String() string {
	k.validate()
	return [...]string{"CursorUp", "CursorDown", "PageUp", "PageDown", "HalfPageUp", "HalfPageDown", "Enter", "Create", "Delete", "Cancel", "Quit", "ForceQuit"}[k]
}

func (k defaultKey) KeyStroke() string {
	k.validate()
	return [...]string{"up", "down", "pgup", "pgdown", "u", "d", "enter", "a", "ctrl+d", "esc", "q", "ctrl+c"}[k]
}

func (k defaultKey) Help() string {
	k.validate()
	return [...]string{
		"Move up",
		"Move down",
		"Page up",
		"Page down",
		"½ page up",
		"½ page down",
		"Select",
		"Create",
		"Delete",
		"Cancel",
		"Quit",
		"Force quit",
	}[k]
}

func (k defaultKey) ShortHelp() string {
	return k.Help()
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
	items     map[KeyProvider]key.Binding
	shortHelp []key.Binding
}

func NewKeyMap() *KeyMap {
	return &KeyMap{
		items:     make(map[KeyProvider]key.Binding),
		shortHelp: make([]key.Binding, 0),
	}
}

func NewListKeyMap() *KeyMap {
	return NewKeyMap().
		WithKey(Up, false).
		WithKey(Down, false).
		WithKey(Enter, true).
		WithKey(Quit, false)
}

func NewViewportKeyMap() *KeyMap {
	return NewKeyMap().
		WithKey(Up, false).
		WithKey(Down, false).
		WithKey(PageUp, false).
		WithKey(PageDown, false).
		WithKey(HalfPageUp, false).
		WithKey(HalfPageDown, false).
		WithKey(Enter, true).
		WithKey(Quit, false).
		WithKey(Cancel, false)
}

func (k *KeyMap) ShortHelp() []key.Binding {
	return k.shortHelp
}

func (k *KeyMap) FullHelp() [][]key.Binding {
	inner := make([]key.Binding, len(k.items))
	i := 0
	for _, v := range k.shortHelp {
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

func (k *KeyMap) Matches(msg tea.KeyMsg, keyProvider KeyProvider) bool {
	return key.Matches(msg, k.items[keyProvider])
}

func (k *KeyMap) WithKey(keyProvider KeyProvider, showShortHelp bool) *KeyMap {
	k.items[keyProvider] = key.NewBinding(
		key.WithKeys(keyProvider.KeyStroke()),
		key.WithHelp(keyProvider.KeyStroke(), keyProvider.Help()),
	)
	if showShortHelp {
		b := key.NewBinding(
			key.WithKeys(keyProvider.KeyStroke()),
			key.WithHelp(keyProvider.KeyStroke(), keyProvider.ShortHelp()))
		k.shortHelp = append(k.shortHelp, b)
	}
	return k
}
