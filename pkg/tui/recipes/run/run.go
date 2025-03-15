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

package run

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hypershift-community/hyper-console/pkg/env"
	"github.com/hypershift-community/hyper-console/pkg/iter"
	"github.com/hypershift-community/hyper-console/pkg/logging"
	"github.com/hypershift-community/hyper-console/pkg/recipes"
	"github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"
	"github.com/hypershift-community/hyper-console/pkg/taskexec"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/keys"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/navigation"
	"github.com/hypershift-community/hyper-console/pkg/tui/lib/styles"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
// const useHighPerformanceRenderer = false
const defaultBufferSize = 256

var (
	Logger = logging.Logger

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return styles.DefaultStyles().Title.BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()

	currentCmdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC66"))
	checkMark       = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

type ExecutionReady int

type RecipeExecuted string

type CommandExecuted string

type CmdOutput struct {
	cmd    string
	output string
	done   bool
}

type bufferReadMsg []byte

//type SafeStack struct {
//	mu    sync.Mutex
//	stack []*CmdOutput
//}
//
//func (s *SafeStack) Push(c *CmdOutput) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	s.stack = append(s.stack, c)
//}
//
//func (s *SafeStack) Pop() (*CmdOutput, bool) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	if len(s.stack) == 0 {
//		return nil, false
//	}
//	val := s.stack[len(s.stack)-1]
//	s.stack = s.stack[:len(s.stack)-1]
//	return val, true
//}
//
//func (s *SafeStack) Len() int {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	return len(s.stack)
//}
//
//func (s *SafeStack) Peek() (*CmdOutput, bool) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	if len(s.stack) == 0 {
//		return nil, false
//	}
//	return s.stack[len(s.stack)-1], true
//}

type model struct {
	recipe          recipes.Recipe
	ready           bool
	width           int
	height          int
	viewport        viewport.Model
	error           error
	keyMap          *keys.KeyMap
	envDir          string
	task            *ast.Task
	execIterator    iter.Iterable[taskexec.Executor]
	index           int
	total           int
	done            bool
	spinner         spinner.Model
	progress        progress.Model
	content         string
	currentCommand  atomic.Pointer[CmdOutput]
	doneView        string
	inProgressView  string
	progressBarView string
	footer          string
	taskReader      io.Reader
	detached        bool
}

func New(width, height int, recipe recipes.Recipe, envDir string) tea.Model {
	m := model{
		recipe: recipe,
		width:  width,
		height: height,
		keyMap: keys.NewViewportKeyMap(),
		envDir: envDir,
	}
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	m.viewport = viewport.New(width, height-verticalMarginHeight)
	m.viewport.YPosition = headerHeight + 10

	m.spinner = spinner.New()
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	m.progress = progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(20),
		progress.WithoutPercentage(),
	)

	return &m
}

func (m *model) Init() tea.Cmd {
	return tea.Sequence(m.prepareExecution(), m.spinner.Tick)
}

func (m *model) prepareExecution() tea.Cmd {
	return func() tea.Msg {
		pr, pw := io.Pipe()
		m.taskReader = pr
		var stdIn bytes.Buffer
		var stdErr bytes.Buffer

		options := []taskexec.TaskOption{taskexec.WithIO(&stdIn, pw, &stdErr)}

		if m.recipe.Environment != "" {
			e, err := env.Load(filepath.Join(m.envDir, m.recipe.Environment))
			if err != nil {
				m.error = err
				return RecipeExecuted("Error loading environment: " + err.Error())

			}
			options = append(options, taskexec.WithEnv(e))
		}
		iter, n, err := taskexec.NewExecutorIterator(m.recipe.Dir, options...)
		if err != nil {
			m.error = err
			return RecipeExecuted("Error setting up recipe executor: " + err.Error())
		}
		m.execIterator = iter
		m.total = n
		m.task = iter.GetTask()
		return ExecutionReady(n)
	}
}

func (m *model) NextCommand() tea.Cmd {

	return func() tea.Msg {
		if !m.execIterator.HasNext() {
			return RecipeExecuted("Recipe executed successfully")
		}
		m.currentCommand.Store(
			&CmdOutput{
				cmd: m.task.Cmds[m.index].Cmd,
			},
		)
		e, err := m.execIterator.Next()
		if err != nil {
			m.error = err
			return RecipeExecuted("Error running recipe: " + err.Error())
		}

		if err := e.Execute(); err != nil {
			m.error = err
			return RecipeExecuted("Error running recipe: " + err.Error())
		}

		return CommandExecuted("Command executed successfully")
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	Logger.Debug("Update :: Received message", "msg", msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case m.keyMap.Matches(msg, keys.Quit) || m.keyMap.Matches(msg, keys.ForceQuit):
			return m, tea.Quit
		case m.keyMap.Matches(msg, keys.Cancel):
			return m, navigation.Back()
		case m.keyMap.Matches(msg, keys.Up) || m.keyMap.Matches(msg, keys.Down) || m.keyMap.Matches(msg, keys.PageUp) || m.keyMap.Matches(msg, keys.PageDown) || m.keyMap.Matches(msg, keys.HalfPageUp) || m.keyMap.Matches(msg, keys.HalfPageDown):
			m.detached = true
		// TODO: Add support for attached mode
		case msg.String() == "a":
			m.detached = false
		default:
			fmt.Println("Key not recognized")
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

	case ExecutionReady:
		m.ready = true
		progressCmd := m.progress.SetPercent(0)
		cmds = append(cmds, tea.Batch(m.NextCommand(), progressCmd))
	case RecipeExecuted:
		m.footer = string(msg)
		m.done = true
	case CommandExecuted:
		m.currentCommand.Load().done = true
		if m.index < m.total {
			m.index++
			progressCmd := m.progress.SetPercent(float64(m.index) / float64(m.total))
			cmds = append(cmds, tea.Batch(m.NextCommand(), progressCmd))
		}
	case spinner.TickMsg:
		if !m.done {
			readMsg := func() tea.Msg {
				buffer := make([]byte, defaultBufferSize)
				n, err := m.taskReader.Read(buffer)
				Logger.Debug("Update :: Received buffer read message", "n", n, "err", err)
				if err != nil {
					if err != io.EOF {
						m.error = err
					}
				}
				return bufferReadMsg(buffer[:n])
			}

			//buffer := make([]byte, defaultBufferSize)
			//n, err := m.taskReader.Read(buffer)
			//Logger.Debug("Update :: Received buffer read message", "n", n, "err", err)
			//if err != nil {
			//	if err != io.EOF {
			//		m.error = err
			//	}
			//}
			//m.lastCommand.output += string(buffer[:n])
			m.spinner, cmd = m.spinner.Update(msg)
			return m, tea.Batch(cmd, readMsg)
		}
	case bufferReadMsg:
		buffer := msg
		n := len(buffer)
		m.currentCommand.Load().output += string(buffer[:n])
	case progress.FrameMsg:
		//if !m.done {
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
		//}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	m.updateVPContent()
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	//return fmt.Sprintf("%s\n%s", m.headerView(), m.content)
}

func (m *model) updateVPContent() {
	m.updateDoneView()
	m.updateInProgressView()
	m.updateProgressBarView()
	log := ""
	if m.doneView != "" {
		log = m.doneView

	}
	if m.inProgressView != "" {
		log += m.inProgressView

	}
	m.content = lipgloss.JoinVertical(lipgloss.Left, log, m.progressBarView, m.footer)
	m.viewport.SetContent(m.content)
	if !m.detached {
		m.viewport.GotoBottom()
	}
}

func (m *model) updateDoneView() {
	cmd := m.currentCommand.Load()
	if cmd == nil {
		return
	}
	sb := strings.Builder{}
	sb.WriteString(m.doneView)
	if cmd.done {
		sb.WriteString(m.inProgressView)
		sb.WriteString(cmd.output)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s Done.\n", checkMark))
		sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("─", m.width)))
		m.currentCommand.Store(nil)
		m.inProgressView = ""
	}
	m.doneView = sb.String()
}

func (m *model) updateInProgressView() {
	cmd := m.currentCommand.Load()
	if cmd == nil {
		return
	}
	sb := strings.Builder{}
	if m.inProgressView == "" {
		sb.WriteString("- " + currentCmdStyle.Render(cmd.cmd) + "\n\n")
	}
	sb.WriteString(m.inProgressView)
	sb.WriteString(cmd.output)
	cmd.output = ""
	m.inProgressView = sb.String()
}
func (m *model) updateProgressBarView() {
	//cmd := m.currentCommand.Load()
	//if cmd == nil {
	//	return
	//}
	n := m.total
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	if m.index == m.total {
		spin = checkMark.String() + " "
	}
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	recipeName := m.recipe.Name
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Executing " + recipeName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	m.progressBarView = spin + info + gap + prog + pkgCount
}

func (m *model) headerView() string {
	title := titleStyle.Render(m.recipe.DisplayName)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
