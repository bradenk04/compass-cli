package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bradenkennedy.com/compass-cli/internal/mongo"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson"
)

type AggregationLoadedMsg struct {
	Docs []bson.M
}

type AggregationErrorMsg struct {
	Err error
}

type AggregationModel struct {
	client   *mongo.Client
	dbName   string
	collName string
	editor   textarea.Model
	viewport viewport.Model
	docs     []bson.M
	loading  bool
	err      error
	width    int
	height   int
	focused  bool
}

func NewAggregationModel() AggregationModel {
	ta := textarea.New()
	ta.Placeholder = `[\n  {\n    "$match": {}\n  }\n]`
	ta.SetValue("[\n  \n]")
	ta.SetHeight(15)
	ta.SetWidth(30)
	ta.Focus()

	vp := viewport.New(0, 0)
	vp.SetContent(" Run pipeline to see results...")

	return AggregationModel{
		editor:   ta,
		viewport: vp,
	}
}

func (m *AggregationModel) SetCollection(client *mongo.Client, db, coll string) {
	m.client = client
	m.dbName = db
	m.collName = coll
	m.docs = nil
	m.err = nil
	m.viewport.SetContent(" Run pipeline to see results...")
}

func (m AggregationModel) RunPipelineCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pipelineStr := m.editor.Value()
		if strings.TrimSpace(pipelineStr) == "" || pipelineStr == "[\n  \n]" {
			pipelineStr = "[]"
		}

		docs, err := m.client.RunAggregation(ctx, m.dbName, m.collName, pipelineStr)
		if err != nil {
			return AggregationErrorMsg{Err: err}
		}
		return AggregationLoadedMsg{Docs: docs}
	}
}

func (m AggregationModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m AggregationModel) Update(msg tea.Msg) (AggregationModel, tea.Cmd) {
	if !m.focused {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, nil
		}
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case AggregationLoadedMsg:
		m.loading = false
		m.docs = msg.Docs
		m.err = nil
		m.updateViewportContent()

	case AggregationErrorMsg:
		m.loading = false
		m.err = msg.Err
		m.viewport.SetContent(ErrorStyle.Render("Aggregation failed: " + msg.Err.Error()))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			m.loading = true
			m.viewport.SetContent(" Running pipeline...")
			return m, m.RunPipelineCmd()
		case "pgup", "ctrl+u":
			m.viewport.LineUp(5)
			return m, nil
		case "pgdown", "ctrl+d":
			m.viewport.LineDown(5)
			return m, nil
		}
	}

	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m *AggregationModel) updateViewportContent() {
	if len(m.docs) == 0 {
		m.viewport.SetContent(" Pipeline returned 0 documents.")
		return
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render(fmt.Sprintf(" RESULTS  ·  %d documents", len(m.docs))) + "\n\n")

	for i, doc := range m.docs {
		bz, err := bson.MarshalExtJSONIndent(doc, true, true, "", "  ")
		if err != nil {
			b.WriteString(fmt.Sprintf("/* error marshaling document %d: %s */\n", i+1, err.Error()))
			continue
		}
		b.WriteString(fmt.Sprintf("/* %d */\n%s\n\n", i+1, HighlightJSON(string(bz))))
	}

	m.viewport.SetContent(b.String())
}

func (m *AggregationModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	leftWidth := width / 2
	rightWidth := width - leftWidth - 2
	contentHeight := height - 5

	m.editor.SetWidth(leftWidth - 2)
	m.editor.SetHeight(contentHeight - 2)
	m.viewport.Width = rightWidth
	m.viewport.Height = contentHeight
}

func (m AggregationModel) View() string {
	leftWidth := m.width / 2

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				TitleStyle.Render(" PIPELINE EDITOR"),
				lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render(" JSON array of stages"),
				"\n",
				m.editor.View(),
			),
		)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, m.viewport.View())

	helpText := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(m.width - 2).
		Render("  " + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render("[ctrl+r] Run  [ctrl+d/u] Scroll results"))

	return ContentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, panes, helpText),
	)
}
