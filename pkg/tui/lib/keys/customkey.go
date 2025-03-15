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

type CustomKey struct {
	name      string
	keystroke string
	help      string
}

// NewCustomKey creates a new custom key with the given name, keystroke and help message
// e.g. NewCustomKey("Reload", "ctrl+r", "Reload the data")
func NewCustomKey(name, keystroke, help string) KeyProvider {
	return &CustomKey{
		name:      name,
		keystroke: keystroke,
		help:      help,
	}
}

func (k *CustomKey) String() string {
	return k.name
}

func (k *CustomKey) KeyStroke() string {
	return k.keystroke
}

func (k *CustomKey) Help() string {
	return k.help
}

func (k *CustomKey) ShortHelp() string {
	return k.name
}
