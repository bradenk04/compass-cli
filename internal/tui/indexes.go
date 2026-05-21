package tui

import (
	"context"
	"time"

	"bradenkennedy.com/compass-cli/internal/mongo"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type IndexesLoadedMsg struct {
	Indexes []mongo.IndexInfo
}

type IndexesErrorMsg struct {
	Err error
}

type IndexesModel struct {
	client   *mongo.Client
	dbName   string
	collName string
	indexes  []mongo.IndexInfo
	table    table.Model
	loading  bool
	err      error
	width    int
	height   int
	focused  bool
}

func NewIndexesModel() IndexesModel {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Keys", Width: 35},
		{Title: "Unique", Width: 8},
		{Title: "Sparse", Width: 8},
		{Title: "Details", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color(ColorGreen))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(ColorSlateBg)).
		Background(lipgloss.Color(ColorGreen)).
		Bold(true)
	t.SetStyles(s)

	return IndexesModel{table: t}
}

func (m *IndexesModel) SetCollection(client *mongo.Client, db, coll string) tea.Cmd {
	m.client = client
	m.dbName = db
	m.collName = coll
	m.indexes = nil
	m.err = nil
	m.loading = true
	return m.FetchIndexesCmd()
}

func (m IndexesModel) FetchIndexesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		indexes, err := m.client.ListIndexes(ctx, m.dbName, m.collName)
		if err != nil {
			return IndexesErrorMsg{Err: err}
		}
		return IndexesLoadedMsg{Indexes: indexes}
	}
}

func (m IndexesModel) Init() tea.Cmd {
	return nil
}

func (m IndexesModel) Update(msg tea.Msg) (IndexesModel, tea.Cmd) {
	if !m.focused {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, nil
		}
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case IndexesLoadedMsg:
		m.loading = false
		m.indexes = msg.Indexes
		m.err = nil

		var rows []table.Row
		for _, idx := range msg.Indexes {
			unique := boolStr(idx.Unique)
			sparse := boolStr(idx.Sparse)
			rows = append(rows, table.Row{idx.Name, idx.Key, unique, sparse, idx.ExtraInfo})
		}
		m.table.SetRows(rows)

	case IndexesErrorMsg:
		m.loading = false
		m.err = msg.Err

	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			m.err = nil
			return m, m.FetchIndexesCmd()
		}
	}

	if !m.loading {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

func (m *IndexesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.table.SetWidth(width - 4)
	m.table.SetHeight(height - 8)
}

func (m IndexesModel) View() string {
	var body string
	switch {
	case m.loading:
		body = "\n  Loading indexes..."
	case m.err != nil:
		body = "\n  " + ErrorStyle.Render("Error: "+m.err.Error())
	case len(m.indexes) == 0:
		body = "\n  No indexes found."
	default:
		body = "\n" + m.table.View()
	}

	helpText := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(m.width - 2).
		Render("  " + lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render("[r] Reload  [↑/↓] Navigate"))

	return ContentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render(" INDEXES"),
			body,
			helpText,
		),
	)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
