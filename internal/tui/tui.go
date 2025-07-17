package tui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/promptkit/promptkit/internal/list"
	"github.com/promptkit/promptkit/pkg/session"
)

func getPathValue(data any, path []string) (any, bool) {
	if len(path) == 0 {
		return data, true
	}
	m, ok := data.(map[string]any)
	if !ok {
		return nil, false
	}
	v, ok := m[path[0]]
	if !ok {
		return nil, false
	}
	return getPathValue(v, path[1:])
}

// Model represents the TUI state.
type Model struct {
	addr     string
	sessions []list.Summary
	cursor   int
	selected map[int]struct{}
	details  *session.Session
	width    int
	height   int
	events   chan session.Session
	viewport viewport.Model
	help     help.Model
	keys     keyMap
	showHelp bool
}

func New(addr string) *Model {
	m := &Model{addr: addr, selected: make(map[int]struct{})}
	m.help = help.New()
	m.keys = newKeyMap()
	return m
}

func (m *Model) Init() tea.Cmd {
	m.viewport = viewport.New(0, 0)
	return tea.Batch(m.loadSessions(), subscribeCmd(m.addr))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if m.cursor >= 0 && m.cursor < len(m.sessions) {
				return m, m.loadDetail(m.sessions[m.cursor].ID)
			}
		case key.Matches(msg, m.keys.Space):
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		default:
			if m.details != nil {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case sessionsMsg:
		m.sessions = msg
		if m.cursor >= len(m.sessions) {
			m.cursor = len(m.sessions) - 1
		}
	case sessionMsg:
		m.sessions = append([]list.Summary{list.Summarize(msg.Session)}, m.sessions...)
	case detailMsg:
		m.details = &msg.Session
		m.viewport.SetContent(renderPrompt(msg.Session))
		m.viewport.GotoTop()
	case subscribeReadyMsg:
		m.events = msg.ch
		return m, waitEventCmd(m.events)
	case eventMsg:
		m.sessions = append([]list.Summary{list.Summarize(msg.Session)}, m.sessions...)
		return m, waitEventCmd(m.events)
	case errorMsg:
		fmt.Println("error:", msg.err)
	}
	return m, nil
}

func (m *Model) View() string {
	leftWidth := 30
	if leftWidth >= m.width {
		leftWidth = m.width - 1
	}
	rightWidth := m.width - leftWidth - 1
	listView := m.renderList(leftWidth)
	detailView := m.renderDetail(rightWidth)
	header := lipgloss.NewStyle().Bold(true).Render("PromptKit UI")
	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
	m.help.ShowAll = m.showHelp
	helpView := m.help.View(m.keys)
	return header + "\n" + body + "\n" + helpView
}

func (m *Model) renderList(width int) string {
	var b strings.Builder
	style := lipgloss.NewStyle().Width(width)
	for i, s := range m.sessions {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		mark := "[ ]"
		if _, ok := m.selected[i]; ok {
			mark = "[x]"
		}
		b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, mark, s.ID))
	}
	return style.Render(b.String())
}

func (m *Model) renderDetail(width int) string {
	if m.details == nil {
		return lipgloss.NewStyle().Width(width).Render("no session selected")
	}
	s := m.details
	smap := list.ToMap(*s)
	modelVal, _ := getPathValue(smap, []string{"request", "model"})
	if modelVal == nil {
		modelVal, _ = getPathValue(smap, []string{"request", "payload", "model"})
	}

	head := fmt.Sprintf("Origin: %s\nModel: %v\nLatency: %dms\nTags: %v", s.Origin, modelVal, s.Metadata.LatencyMS, s.Metadata.Tags)

	helpHeight := lipgloss.Height(m.help.View(m.keys))
	headerLines := 1
	metaLines := 4 // origin + model + latency + tags
	vh := m.height - headerLines - metaLines - helpHeight
	if vh < 1 {
		vh = 1
	}
	m.viewport.Width = width
	m.viewport.Height = vh

	return lipgloss.NewStyle().Width(width).Render(head + "\n" + strings.Repeat("-", width) + "\n" + m.viewport.View())
}

func renderPrompt(s session.Session) string {
	var b strings.Builder
	if sp := strings.TrimSpace(s.SourcePrompt); sp != "" {
		fmt.Fprintf(&b, "Prompt:\n%s\n", sp)
	}
	if resp, err := json.MarshalIndent(s.Response, "", "  "); err == nil {
		fmt.Fprintf(&b, "\nResponse:\n%s\n", resp)
	}
	return b.String()
}

// keyMap defines key bindings for the UI.
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Space key.Binding
	Help  key.Binding
	Quit  key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "details")),
		Space: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "select")),
		Help:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Space, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Enter, k.Space, k.Help, k.Quit}}
}

// messages

type sessionsMsg []list.Summary

type sessionMsg struct{ Session session.Session }

type detailMsg struct{ Session session.Session }

type subscribeReadyMsg struct{ ch chan session.Session }

type eventMsg struct{ Session session.Session }

type errorMsg struct{ err error }

// commands

func (m *Model) loadSessions() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://" + m.addr + "/sessions")
		if err != nil {
			return errorMsg{err}
		}
		defer resp.Body.Close()
		var out []list.Summary
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return errorMsg{err}
		}
		return sessionsMsg(out)
	}
}

func (m *Model) loadDetail(id string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://" + m.addr + "/sessions/" + id)
		if err != nil {
			return errorMsg{err}
		}
		defer resp.Body.Close()
		var sess session.Session
		if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
			return errorMsg{err}
		}
		return detailMsg{sess}
	}
}

func subscribeCmd(addr string) tea.Cmd {
	return func() tea.Msg {
		url := "http://" + addr + "/events"
		resp, err := http.Get(url)
		if err != nil {
			return errorMsg{err}
		}
		ch := make(chan session.Session)
		go func() {
			defer resp.Body.Close()
			reader := bufio.NewReader(resp.Body)
			var data bytes.Buffer
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					close(ch)
					return
				}
				if bytes.Equal(line, []byte("\n")) {
					if data.Len() > 0 {
						var sess session.Session
						if err := json.Unmarshal(data.Bytes(), &sess); err == nil {
							ch <- sess
						}
						data.Reset()
					}
					continue
				}
				if bytes.HasPrefix(line, []byte("data:")) {
					data.Write(bytes.TrimSpace(line[len("data:"):]))
				}
			}
		}()
		return subscribeReadyMsg{ch}
	}
}

func waitEventCmd(ch chan session.Session) tea.Cmd {
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return errorMsg{fmt.Errorf("event stream closed")}
		}
		return eventMsg{s}
	}
}
