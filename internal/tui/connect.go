package tui

import (
	"fmt"

	"bradenkennedy.com/compass-cli/internal/mongo"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConnectedMsg struct {
	Client *mongo.Client
}

type ConnectErrMsg struct {
	Err error
}

type ConnectModel struct {
	textInput     textinput.Model
	spinner       spinner.Model
	styles        Styles
	loading       bool
	err           error
	width         int
	height        int
	history       []string
	historyCursor int
}

func NewConnectModel(styles Styles) ConnectModel {
	ti := textinput.New()
	ti.Placeholder = "mongodb://localhost:27017"
	ti.SetValue("mongodb://localhost:27017")
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.LabelStyle

	history := LoadHistory()

	if len(history) > 0 {
		ti.SetValue(history[0])
	}

	return ConnectModel{
		textInput:     ti,
		spinner:       s,
		styles:        styles,
		loading:       false,
		history:       history,
		historyCursor: 0,
	}
}

func (m ConnectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ConnectModel) Update(msg tea.Msg) (ConnectModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "enter":
			uri := m.textInput.Value()
			if uri == "" {
				uri = "mongodb://localhost:27017"
			}
			m.loading = true
			m.err = nil
			return m, tea.Batch(
				m.spinner.Tick,
				connectCmd(uri),
			)
		case "down":
			if len(m.history) > 0 {
				if m.historyCursor < len(m.history)-1 {
					m.historyCursor++
					m.textInput.SetValue(m.history[m.historyCursor])
				}
				return m, nil
			}
		case "up":
			if len(m.history) > 0 {
				if m.historyCursor > -1 {
					m.historyCursor--
					if m.historyCursor == -1 {
						m.textInput.SetValue("mongodb://localhost:27017")
					} else {
						m.textInput.SetValue(m.history[m.historyCursor])
					}
				}
				return m, nil
			}
		default:
			m.historyCursor = -1
		}

	case ConnectedMsg:
		m.loading = false
		m.err = nil

	case ConnectErrMsg:
		m.loading = false
		m.err = msg.Err

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func connectCmd(uri string) tea.Cmd {
	return func() tea.Msg {
		client, err := mongo.Connect(uri)
		if err != nil {
			return ConnectErrMsg{Err: err}
		}
		return ConnectedMsg{Client: client}
	}
}

func (m ConnectModel) View() string {
	cardContent := ""

	title := m.styles.TitleStyle.Render("MONGODB COMPASS CLI")
	subtitle := m.styles.SubtitleStyle.Render("Interactive Terminal MongoDB Client")

	inputLabel := m.styles.InputLabelStyle.Render("Connection string (URI):")

	inputView := m.textInput.View()

	var statusView string
	if m.loading {
		statusView = fmt.Sprintf("\n  %s Connecting to database...", m.spinner.View())
	} else if m.err != nil {
		statusView = fmt.Sprintf("\n  %s %s", m.styles.ErrorStyle.Render("Connection failed:"), m.styles.ErrorStyle.Render(m.err.Error()))
	} else {
		statusView = "\n  Press Enter to connect. Esc to exit."
	}

	var historyView string
	if len(m.history) > 0 {
		historyView = "\n" + m.styles.TextMutedStyle.Bold(true).Render("Recent Connections (use Up/Down to select):") + "\n"
		for i, h := range m.history {
			indicator := "  "
			style := m.styles.TextMutedStyle
			if i == m.historyCursor {
				indicator = "➔ "
				style = m.styles.TextMutedStyle.Bold(true)
			}
			displayUri := h
			if len(displayUri) > 54 {
				displayUri = displayUri[:51] + "..."
			}
			historyView += fmt.Sprintf("%s%s\n", indicator, style.Render(displayUri))
		}
	}

	cardContent = fmt.Sprintf(
		"%s\n%s\n\n%s\n%s%s\n%s",
		title,
		subtitle,
		inputLabel,
		inputView,
		historyView,
		statusView,
	)

	card := m.styles.CardStyle.
		Width(60).
		Align(lipgloss.Left).
		Render(cardContent)

	vOffset := (m.height - lipgloss.Height(card)) / 2
	hOffset := (m.width - lipgloss.Width(card)) / 2

	if vOffset < 0 {
		vOffset = 0
	}
	if hOffset < 0 {
		hOffset = 0
	}

	return lipgloss.NewStyle().
		PaddingTop(vOffset).
		PaddingLeft(hOffset).
		Width(m.width).
		Height(m.height).
		Render(card)
}
