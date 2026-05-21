package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"bradenkennedy.com/compass-cli/internal/mongo"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type docMode int

const (
	docModeList docMode = iota
	docModeEdit
)

type DocsLoadedMsg struct {
	Docs  []bson.M
	Total int64
}

type DocsErrorMsg struct {
	Err error
}

type DocSavedMsg struct{}

type DocumentsModel struct {
	client    *mongo.Client
	dbName    string
	collName  string
	docs      []bson.M
	totalDocs int64
	cursor    int
	mode      docMode

	filterInput textinput.Model
	sortInput   textinput.Model
	limitInput  textinput.Model
	skipInput   textinput.Model
	activeInput int

	editor    textarea.Model
	editorErr error
	isNewDoc  bool

	viewport viewport.Model
	width    int
	height   int
	focused  bool

	status    string
	statusErr bool
}

func NewDocumentsModel() DocumentsModel {
	filter := textinput.New()
	filter.Prompt = "Filter: "
	filter.Placeholder = "{}"
	filter.SetValue("{}")
	filter.Width = 30

	sort := textinput.New()
	sort.Prompt = "Sort:   "
	sort.Placeholder = `{"_id": -1}`
	sort.SetValue(`{"_id": -1}`)
	sort.Width = 20

	limit := textinput.New()
	limit.Prompt = "Limit:  "
	limit.Placeholder = "20"
	limit.SetValue("20")
	limit.Width = 6

	skip := textinput.New()
	skip.Prompt = "Skip:   "
	skip.Placeholder = "0"
	skip.SetValue("0")
	skip.Width = 6

	ta := textarea.New()
	ta.Placeholder = "{\n  \n}"
	ta.SetHeight(15)
	ta.SetWidth(60)

	return DocumentsModel{
		docs:        []bson.M{},
		filterInput: filter,
		sortInput:   sort,
		limitInput:  limit,
		skipInput:   skip,
		activeInput: -1,
		editor:      ta,
		viewport:    viewport.New(0, 0),
		mode:        docModeList,
	}
}

func (m *DocumentsModel) SetCollection(client *mongo.Client, db, coll string) tea.Cmd {
	m.client = client
	m.dbName = db
	m.collName = coll
	m.cursor = 0
	m.docs = []bson.M{}
	m.status = ""
	m.statusErr = false
	return m.FetchDocsCmd()
}

func (m DocumentsModel) FetchDocsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		limit, _ := strconv.ParseInt(m.limitInput.Value(), 10, 64)
		if limit <= 0 {
			limit = 20
		}
		skip, _ := strconv.ParseInt(m.skipInput.Value(), 10, 64)

		opts := mongo.QueryOptions{
			Filter: m.filterInput.Value(),
			Sort:   m.sortInput.Value(),
			Limit:  limit,
			Skip:   skip,
		}

		docs, total, err := m.client.FetchDocuments(ctx, m.dbName, m.collName, opts)
		if err != nil {
			return DocsErrorMsg{Err: err}
		}
		return DocsLoadedMsg{Docs: docs, Total: total}
	}
}

func (m DocumentsModel) deleteDocCmd(id interface{}) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.client.DeleteDocument(ctx, m.dbName, m.collName, id); err != nil {
			return DocsErrorMsg{Err: err}
		}
		return DocSavedMsg{}
	}
}

func (m DocumentsModel) saveDocCmd(id interface{}, doc bson.M, isNew bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		if isNew {
			_, err = m.client.InsertDocument(ctx, m.dbName, m.collName, doc)
		} else {
			err = m.client.UpdateDocument(ctx, m.dbName, m.collName, id, doc)
		}

		if err != nil {
			return DocsErrorMsg{Err: err}
		}
		return DocSavedMsg{}
	}
}

func (m DocumentsModel) Init() tea.Cmd {
	return nil
}

func (m DocumentsModel) Update(msg tea.Msg) (DocumentsModel, tea.Cmd) {
	if !m.focused {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, nil
		}
	}

	var cmd tea.Cmd

	switch m.mode {
	case docModeList:
		if m.activeInput != -1 {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.Type {
				case tea.KeyEsc:
					m.blurAllInputs()
					return m, nil
				case tea.KeyEnter:
					m.blurAllInputs()
					m.cursor = 0
					m.status = "Querying..."
					m.statusErr = false
					return m, m.FetchDocsCmd()
				case tea.KeyTab:
					m.nextInput()
					return m, nil
				}
			}

			switch m.activeInput {
			case 0:
				m.filterInput, cmd = m.filterInput.Update(msg)
			case 1:
				m.sortInput, cmd = m.sortInput.Update(msg)
			case 2:
				m.limitInput, cmd = m.limitInput.Update(msg)
			case 3:
				m.skipInput, cmd = m.skipInput.Update(msg)
			}
			return m, cmd
		}

		switch msg := msg.(type) {
		case DocsLoadedMsg:
			m.docs = msg.Docs
			m.totalDocs = msg.Total
			m.status = fmt.Sprintf("Found %d documents", m.totalDocs)
			m.statusErr = false
			if m.cursor >= len(m.docs) {
				m.cursor = max(0, len(m.docs)-1)
			}
			m.updateViewportContent()

		case DocsErrorMsg:
			m.status = msg.Err.Error()
			m.statusErr = true

		case DocSavedMsg:
			m.status = "Document saved successfully"
			m.statusErr = false
			return m, m.FetchDocsCmd()

		case tea.KeyMsg:
			switch msg.String() {
			case "/":
				m.activeInput = 0
				m.filterInput.Focus()
				m.sortInput.Blur()
				m.limitInput.Blur()
				m.skipInput.Blur()
			case "s":
				m.activeInput = 1
				m.filterInput.Blur()
				m.sortInput.Focus()
				m.limitInput.Blur()
				m.skipInput.Blur()
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.updateViewportContent()
				}
			case "down", "j":
				if m.cursor < len(m.docs)-1 {
					m.cursor++
					m.updateViewportContent()
				}
			case "ctrl+d", "pgdown":
				m.viewport.LineDown(3)
			case "ctrl+u", "pgup":
				m.viewport.LineUp(3)
			case "r":
				m.status = "Refreshing..."
				return m, m.FetchDocsCmd()
			case "i":
				m.mode = docModeEdit
				m.isNewDoc = true
				m.editor.SetValue("{\n  \n}")
				m.editor.Focus()
				m.editorErr = nil
				return m, textarea.Blink
			case "e", "enter":
				if len(m.docs) > 0 {
					m.mode = docModeEdit
					m.isNewDoc = false
					bz, _ := bson.MarshalExtJSONIndent(m.docs[m.cursor], true, true, "", "  ")
					m.editor.SetValue(string(bz))
					m.editor.Focus()
					m.editorErr = nil
					return m, textarea.Blink
				}
			case "d", "x":
				if len(m.docs) > 0 {
					return m, m.deleteDocCmd(m.docs[m.cursor]["_id"])
				}
			}
		}

	case docModeEdit:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				m.mode = docModeList
				m.editor.Blur()
				m.updateViewportContent()
				return m, nil
			case "ctrl+s":
				var doc bson.M
				err := bson.UnmarshalExtJSON([]byte(m.editor.Value()), true, &doc)
				if err != nil {
					m.editorErr = fmt.Errorf("JSON syntax error: %w", err)
					return m, nil
				}

				m.mode = docModeList
				m.editor.Blur()

				var id interface{}
				if !m.isNewDoc {
					id = m.docs[m.cursor]["_id"]
					if _, hasID := doc["_id"]; !hasID {
						doc["_id"] = id
					}
				}
				return m, m.saveDocCmd(id, doc, m.isNewDoc)
			}
		}

		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *DocumentsModel) blurAllInputs() {
	m.activeInput = -1
	m.filterInput.Blur()
	m.sortInput.Blur()
	m.limitInput.Blur()
	m.skipInput.Blur()
}

func (m *DocumentsModel) nextInput() {
	m.activeInput = (m.activeInput + 1) % 4
	m.filterInput.Blur()
	m.sortInput.Blur()
	m.limitInput.Blur()
	m.skipInput.Blur()

	switch m.activeInput {
	case 0:
		m.filterInput.Focus()
	case 1:
		m.sortInput.Focus()
	case 2:
		m.limitInput.Focus()
	case 3:
		m.skipInput.Focus()
	}
}

func (m *DocumentsModel) updateViewportContent() {
	if len(m.docs) == 0 {
		m.viewport.SetContent(" No documents found.")
		return
	}

	bz, err := bson.MarshalExtJSONIndent(m.docs[m.cursor], true, true, "", "  ")
	if err != nil {
		m.viewport.SetContent("Error marshaling document: " + err.Error())
		return
	}

	m.viewport.SetContent(HighlightJSON(string(bz)))
}

func (m *DocumentsModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	listWidth := width / 3
	rightWidth := width - listWidth - 3
	contentHeight := height - 7

	m.viewport.Width = rightWidth
	m.viewport.Height = contentHeight
	m.editor.SetWidth(width - 4)
	m.editor.SetHeight(height - 8)
}

func (m DocumentsModel) View() string {
	if m.mode == docModeEdit {
		header := TitleStyle.Render(" DOCUMENT EDITOR") + "\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).Render(" Ctrl+S to save  ·  Esc to cancel") + "\n\n"

		var errLine string
		if m.editorErr != nil {
			errLine = ErrorStyle.Render(m.editorErr.Error()) + "\n\n"
		}

		return ContentStyle.Render(header + errLine + m.editor.View())
	}

	queryBarContent := lipgloss.JoinHorizontal(lipgloss.Top,
		m.filterInput.View(), "  ",
		m.sortInput.View(), "  ",
		m.limitInput.View(), "  ",
		m.skipInput.View(),
	)

	queryStyle := CardStyle.Width(m.width-4).Padding(0, 1)
	if m.activeInput != -1 {
		queryStyle = CardActiveStyle.Width(m.width-4).Padding(0, 1)
	}
	queryBar := queryStyle.Render(queryBarContent)

	listWidth := m.width / 3
	rightWidth := m.width - listWidth - 4
	contentHeight := m.height - 8

	var listContent string
	if len(m.docs) == 0 {
		listContent = "\n  No documents."
	} else {
		for i, doc := range m.docs {
			label := formatDocLabel(doc)
			if i == m.cursor {
				listContent += CollSelectedStyle.Width(listWidth-2).Render(label) + "\n"
			} else {
				listContent += lipgloss.NewStyle().Width(listWidth-2).Render(label) + "\n"
			}
		}
	}

	listPane := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(listWidth).
		Height(contentHeight).
		Render(listContent)

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight).
		Render(m.viewport.View())

	panes := lipgloss.JoinHorizontal(lipgloss.Top, listPane, rightPane)

	var statusMsg string
	if m.statusErr {
		statusMsg = ErrorStyle.Render(m.status)
	} else {
		statusMsg = SuccessStyle.Render(m.status)
	}

	helpLine := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted)).
		Render("[i] Insert  [e/Enter] Edit  [d] Delete  [r] Refresh  [/] Filter  [s] Sort  [Tab] Cycle inputs  [ctrl+d] Scroll doc")

	statusBar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(m.width - 2).
		Render(fmt.Sprintf("  %s  %s", statusMsg, helpLine))

	return lipgloss.JoinVertical(lipgloss.Left, queryBar, panes, statusBar)
}

func formatDocLabel(doc bson.M) string {
	var label string

	switch v := doc["_id"].(type) {
	case primitive.ObjectID:
		label = fmt.Sprintf("ObjectId(\"%s\")", v.Hex())
	default:
		bz, _ := json.Marshal(doc["_id"])
		label = string(bz)
	}

	if name, ok := doc["name"].(string); ok {
		label += fmt.Sprintf(" (%s)", name)
	} else if title, ok := doc["title"].(string); ok {
		label += fmt.Sprintf(" (%s)", title)
	}

	return " " + label
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
