package tui

import (
	"context"
	"fmt"
	"time"

	"bradenkennedy.com/compass-cli/internal/config"
	"bradenkennedy.com/compass-cli/internal/mongo"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	stateConnect appState = iota
	stateExplorer
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusContent
)

type DBsLoadedMsg struct {
	DBs []mongo.DBInfo
}

type MainModel struct {
	state     appState
	focus     focusArea
	activeTab int
	theme     Styles

	client        *mongo.Client
	connectionURI string
	dbName        string
	collName      string

	connectModel ConnectModel
	sidebarModel SidebarModel
	docsModel    DocumentsModel
	schemaModel  SchemaModel
	aggModel     AggregationModel
	idxModel     IndexesModel

	width  int
	height int
}

func NewMainModel(cfg *config.Config) MainModel {
	styles := InitStyles(cfg.Theme)

	return MainModel{
		state:        stateConnect,
		focus:        focusSidebar,
		activeTab:    0,
		theme:        styles,
		connectModel: NewConnectModel(styles),
		sidebarModel: NewSidebarModel(styles),
		docsModel:    NewDocumentsModel(styles),
		schemaModel:  NewSchemaModel(styles),
		aggModel:     NewAggregationModel(styles),
		idxModel:     NewIndexesModel(styles),
	}
}

func listDBsCmd(client *mongo.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dbs, err := client.ListDatabasesAndCollections(ctx)
		if err != nil {
			return ConnectErrMsg{Err: err}
		}
		return DBsLoadedMsg{DBs: dbs}
	}
}

func (m MainModel) Init() tea.Cmd {
	return m.connectModel.Init()
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.connectModel.width = msg.Width
		m.connectModel.height = msg.Height
		m.recalculateExplorerSizes()

	case ConnectedMsg:
		m.client = msg.Client
		m.connectionURI = m.connectModel.textInput.Value()
		if m.connectionURI == "" {
			m.connectionURI = "mongodb://localhost:27017"
		}
		AddToHistory(m.connectionURI)
		m.state = stateExplorer
		m.focus = focusSidebar
		m.sidebarModel.focused = true
		m.setTabFocus(m.activeTab, false)
		return m, listDBsCmd(m.client)

	case ConnectErrMsg:
		m.connectModel, cmd = m.connectModel.Update(msg)
		return m, cmd

	case DBsLoadedMsg:
		m.sidebarModel.UpdateItems(msg.DBs)
		for _, db := range msg.DBs {
			if len(db.Collections) > 0 {
				m.dbName = db.Name
				m.collName = db.Collections[0]
				m.sidebarModel.selectedDb = db.Name
				m.sidebarModel.selectedColl = db.Collections[0]
				cmds = append(cmds, m.selectCollectionCmd(db.Name, db.Collections[0]))
				break
			}
		}

	case CollectionSelectedMsg:
		m.dbName = msg.Db
		m.collName = msg.Coll
		m.switchTab(0)
		cmds = append(cmds, m.selectCollectionCmd(msg.Db, msg.Coll))

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			if m.client != nil {
				_ = m.client.Disconnect()
			}
			return m, tea.Quit
		}

		if m.state == stateExplorer && !m.isEditingMode() {
			switch msg.String() {
			case "tab":
				if m.focus == focusSidebar {
					m.focus = focusContent
					m.sidebarModel.focused = false
					m.setTabFocus(m.activeTab, true)
				} else {
					m.focus = focusSidebar
					m.sidebarModel.focused = true
					m.setTabFocus(m.activeTab, false)
				}
				return m, nil
			case "1":
				m.switchTab(0)
				return m, nil
			case "2":
				m.switchTab(1)
				return m, nil
			case "3":
				m.switchTab(2)
				return m, nil
			case "4":
				m.switchTab(3)
				return m, nil
			case "esc":
				if m.focus == focusContent {
					m.focus = focusSidebar
					m.sidebarModel.focused = true
					m.setTabFocus(m.activeTab, false)
					return m, nil
				}
				if m.client != nil {
					_ = m.client.Disconnect()
					m.client = nil
				}
				m.state = stateConnect
				m.connectModel = NewConnectModel(m.theme)
				m.connectModel.width = m.width
				m.connectModel.height = m.height
				return m, m.connectModel.Init()
			}
		}
	}

	if m.state == stateConnect {
		m.connectModel, cmd = m.connectModel.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.sidebarModel, cmd = m.sidebarModel.Update(msg)
		cmds = append(cmds, cmd)

		switch msg.(type) {
		case DocsLoadedMsg, DocsErrorMsg, DocSavedMsg:
			m.docsModel, cmd = m.docsModel.Update(msg)
			cmds = append(cmds, cmd)
		case SchemaLoadedMsg, SchemaErrorMsg:
			m.schemaModel, cmd = m.schemaModel.Update(msg)
			cmds = append(cmds, cmd)
		case AggregationLoadedMsg, AggregationErrorMsg:
			m.aggModel, cmd = m.aggModel.Update(msg)
			cmds = append(cmds, cmd)
		case IndexesLoadedMsg, IndexesErrorMsg:
			m.idxModel, cmd = m.idxModel.Update(msg)
			cmds = append(cmds, cmd)
		default:
			switch m.activeTab {
			case 0:
				m.docsModel, cmd = m.docsModel.Update(msg)
				cmds = append(cmds, cmd)
			case 1:
				m.schemaModel, cmd = m.schemaModel.Update(msg)
				cmds = append(cmds, cmd)
			case 2:
				if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
					m.focus = focusSidebar
					m.sidebarModel.focused = true
					m.setTabFocus(m.activeTab, false)
					return m, nil
				}
				m.aggModel, cmd = m.aggModel.Update(msg)
				cmds = append(cmds, cmd)
			case 3:
				m.idxModel, cmd = m.idxModel.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *MainModel) isEditingMode() bool {
	return m.docsModel.mode == docModeEdit || m.docsModel.activeInput != -1 || m.aggModel.focused
}

func (m *MainModel) setTabFocus(tab int, focus bool) {
	m.docsModel.focused = tab == 0 && focus
	m.schemaModel.focused = tab == 1 && focus
	m.aggModel.focused = tab == 2 && focus
	m.idxModel.focused = tab == 3 && focus
}

func (m *MainModel) switchTab(tab int) {
	m.setTabFocus(m.activeTab, false)
	m.activeTab = tab
	m.focus = focusContent
	m.sidebarModel.focused = false
	m.setTabFocus(tab, true)
	m.recalculateExplorerSizes()
}

func (m *MainModel) selectCollectionCmd(db, coll string) tea.Cmd {
	m.aggModel.SetCollection(m.client, db, coll)
	return tea.Batch(
		m.docsModel.SetCollection(m.client, db, coll),
		m.schemaModel.SetCollection(m.client, db, coll),
		m.idxModel.SetCollection(m.client, db, coll),
	)
}

func (m *MainModel) recalculateExplorerSizes() {
	const sidebarWidth = 25
	contentWidth := m.width - sidebarWidth - 1
	contentHeight := m.height - 4

	m.sidebarModel.width = sidebarWidth
	m.sidebarModel.height = m.height - 3

	m.docsModel.SetSize(contentWidth, contentHeight)
	m.schemaModel.SetSize(contentWidth, contentHeight)
	m.aggModel.SetSize(contentWidth, contentHeight)
	m.idxModel.SetSize(contentWidth, contentHeight)
}

func (m MainModel) View() string {
	if m.state == stateConnect {
		return m.connectModel.View()
	}

	header := m.theme.HeaderStyle.Render(fmt.Sprintf(
		"  MONGODB COMPASS CLI  ·  %s  ·  %s.%s",
		m.connectionURI, m.dbName, m.collName,
	))

	tabs := []string{"[1] Documents", "[2] Schema", "[3] Aggregations", "[4] Indexes"}
	var renderedTabs []string
	for i, t := range tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, m.theme.ActiveTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, m.theme.InactiveTabStyle.Render(t))
		}
	}

	tabGapWidth := m.width - 2
	for _, t := range renderedTabs {
		tabGapWidth -= lipgloss.Width(t)
	}
	if tabGapWidth < 0 {
		tabGapWidth = 0
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top,
		renderedTabs[0],
		renderedTabs[1],
		renderedTabs[2],
		renderedTabs[3],
		m.theme.TabGapStyle.Width(tabGapWidth).Render(""),
	)

	var contentView string
	switch m.activeTab {
	case 0:
		contentView = m.docsModel.View()
	case 1:
		contentView = m.schemaModel.View()
	case 2:
		contentView = m.aggModel.View()
	case 3:
		contentView = m.idxModel.View()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarModel.View(), contentView)

	var focusLabel string
	if m.focus == focusSidebar {
		focusLabel = m.theme.LabelStyle.Render("Sidebar")
	} else {
		focusLabel = m.theme.LabelStyle.Render("Content")
	}

	footerStyle := m.theme.FooterStyle.Width(m.width)

	footer := footerStyle.Render(fmt.Sprintf(
		"  Focus: %s  ·  [Tab] Toggle  ·  [1–4] Tabs  ·  [Esc] Back  ·  [Ctrl+C] Quit",
		focusLabel,
	))

	return lipgloss.JoinVertical(lipgloss.Left, header, tabRow, body, footer)
}
