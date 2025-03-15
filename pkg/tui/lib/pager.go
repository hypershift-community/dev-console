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

package lib

import (
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hypershift-community/hyper-console/pkg/logging"
)

const (
	bufferSize   = 256
	tickInterval = 100
)

var Logger = logging.Logger

type bufferReadMsg []byte
type tickMsg time.Time

type model struct {
	viewport *viewport.Model
	width    int
	height   int
	in       io.Reader
	content  strings.Builder
}

func New(width, height int, in io.Reader) tea.Model {
	return &model{
		viewport: &viewport.Model{Width: width, Height: height},
		width:    width,
		height:   height,
		in:       in,
	}
}

func (m *model) Init() tea.Cmd {
	readCmd := func() tea.Msg {
		Logger.Debug("Init :: Reading from input")
		b := make([]byte, bufferSize)
		n, err := m.in.Read(b)
		Logger.Debug("Init :: Read from input", "n", n, "err", err)
		return bufferReadMsg(b[:n])
	}
	return tea.Sequence(readCmd, doTick())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	Logger.Debug("Update :: Received message", "msg", msg)
	switch msg := msg.(type) {
	case bufferReadMsg:
		n := len(msg)
		Logger.Debug("Update :: Received buffer read message", "n", n)
		m.content.Write(msg)
		cmds = append(cmds, m.readNext())
	}
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	Logger.Debug("View :: Rendering view")
	return m.content.String()
}

func (m *model) readNext() tea.Cmd {
	readCmd := func() tea.Msg {
		Logger.Debug("readNext :: Reading from input")
		b := make([]byte, bufferSize)
		n, err := m.in.Read(b)
		Logger.Debug("readNext :: Read from input", "n", n, "err", err)
		return bufferReadMsg(b[:n])
	}
	return tea.Sequence(readCmd, doTick())
}

func doTick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
