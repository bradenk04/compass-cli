package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bradenkennedy.com/compass-cli/internal/mongo"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SchemaLoadedMsg struct {
	Fields []mongo.SchemaField
}

type SchemaErrorMsg struct {
	Err error
}

type SchemaModel struct {
	client   *mongo.Client
	dbName   string
	collName string
	fields   []mongo.SchemaField
	loading  bool
	err      error
	viewport viewport.Model
	width    int
	height   int
	focused  bool
}

func NewSchemaModel() SchemaModel {
	return SchemaModel{
		viewport: viewport.New(0, 0),
	}
}

func (m *SchemaModel) SetCollection(client *mongo.Client, db, coll string) tea.Cmd {
	m.client = client
	m.dbName = db
	m.collName = coll
	m.fields = nil
	m.err = nil
	m.loading = true
	m.viewport.SetContent(" Analyzing schema...")
	return m.AnalyzeSchemaCmd()
}

func (m SchemaModel) AnalyzeSchemaCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fields, err := m.client.AnalyzeSchema(ctx, m.dbName, m.collName)
		if err != nil {
			return SchemaErrorMsg{Err: err}
		}
		return SchemaLoadedMsg{Fields: fields}
	}
}

func (m SchemaModel) Init() tea.Cmd {
	return nil
}

func (m SchemaModel) Update(msg tea.Msg) (SchemaModel, tea.Cmd) {
	if !m.focused {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case SchemaLoadedMsg:
		m.loading = false
		m.fields = msg.Fields
		m.err = nil
		m.updateViewportContent()

	case SchemaErrorMsg:
		m.loading = false
		m.err = msg.Err
		m.viewport.SetContent(ErrorStyle.Render("Schema analysis failed: " + msg.Err.Error()))

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			m.err = nil
			m.viewport.SetContent(" Analyzing schema...")
			return m, m.AnalyzeSchemaCmd()
		case "up", "k":
			m.viewport.LineUp(1)
		case "down", "j":
			m.viewport.LineDown(1)
		case "pgup", "ctrl+u":
			m.viewport.LineUp(5)
		case "pgdown", "ctrl+d":
			m.viewport.LineDown(5)
		}
	}

	return m, nil
}

func (m *SchemaModel) updateViewportContent() {
	if len(m.fields) == 0 {
		m.viewport.SetContent(" No schema fields detected — collection may be empty.")
		return
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render(fmt.Sprintf(" SCHEMA ANALYSIS  ·  sampled %d documents", 100)) + "\n\n")

	for _, f := range m.fields {
		b.WriteString(fmt.Sprintf("%s\n", KeyStyle.Render("• "+f.Path)))

		for typeName, count := range f.Types {
			pct := float64(count) / float64(f.TotalDocCount)

			typeStyle := ValBoolStyle
			switch typeName {
			case "String":
				typeStyle = ValStrStyle
			case "Number", "Double":
				typeStyle = ValNumStyle
			case "Object", "Array":
				typeStyle = InfoStyle
			}

			b.WriteString(fmt.Sprintf(
				"   %s %3d%%  %s  %s\n",
				SuccessStyle.Render(drawProgressBar(pct, 15)),
				int(pct*100),
				typeStyle.Render(typeName),
				lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render(fmt.Sprintf("(%d/%d docs)", count, f.TotalDocCount)),
			))
		}
		b.WriteString("\n")
	}

	m.viewport.SetContent(b.String())
}

func drawProgressBar(pct float64, width int) string {
	filled := int(pct * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func (m *SchemaModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width - 4
	m.viewport.Height = height - 5
}

func (m SchemaModel) View() string {
	helpText := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(m.width - 2).
		Render("  " + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render("[r] Re-analyze  [j/k] Scroll"))

	return ContentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.viewport.View(),
			helpText,
		),
	)
}
