package tui

import (
	"fmt"

	"bradenkennedy.com/compass-cli/internal/mongo"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CollectionSelectedMsg struct {
	Db   string
	Coll string
}

type SidebarItem struct {
	DbName     string
	CollName   string
	IsDb       bool
	IsExpanded bool
}

type SidebarModel struct {
	dbs          []mongo.DBInfo
	flatItems    []SidebarItem
	cursor       int
	selectedDb   string
	selectedColl string
	height       int
	width        int
	focused      bool
}

func NewSidebarModel() SidebarModel {
	return SidebarModel{
		dbs:       []mongo.DBInfo{},
		flatItems: []SidebarItem{},
		cursor:    0,
		focused:   true,
	}
}

func (m *SidebarModel) UpdateItems(dbs []mongo.DBInfo) {
	m.dbs = dbs

	expansionState := make(map[string]bool)
	for _, item := range m.flatItems {
		if item.IsDb {
			expansionState[item.DbName] = item.IsExpanded
		}
	}

	var items []SidebarItem
	for _, db := range dbs {
		expanded, known := expansionState[db.Name]
		if !known {
			expanded = true
		}

		items = append(items, SidebarItem{
			DbName:     db.Name,
			IsDb:       true,
			IsExpanded: expanded,
		})

		if expanded {
			for _, coll := range db.Collections {
				items = append(items, SidebarItem{
					DbName:   db.Name,
					CollName: coll,
					IsDb:     false,
				})
			}
		}
	}

	m.flatItems = items
	if m.cursor >= len(m.flatItems) {
		m.cursor = len(m.flatItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m SidebarModel) Init() tea.Cmd {
	return nil
}

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || len(m.flatItems) == 0 {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.flatItems)-1 {
			m.cursor++
		}
	case "right", " ":
		if item := m.flatItems[m.cursor]; item.IsDb && !item.IsExpanded {
			m.toggleExpand(item.DbName)
		}
	case "left":
		if item := m.flatItems[m.cursor]; item.IsDb && item.IsExpanded {
			m.toggleExpand(item.DbName)
		}
	case "enter":
		item := m.flatItems[m.cursor]
		if item.IsDb {
			m.toggleExpand(item.DbName)
		} else {
			m.selectedDb = item.DbName
			m.selectedColl = item.CollName
			return m, func() tea.Msg {
				return CollectionSelectedMsg{Db: item.DbName, Coll: item.CollName}
			}
		}
	}

	return m, nil
}

func (m *SidebarModel) toggleExpand(dbName string) {
	for i, item := range m.flatItems {
		if item.IsDb && item.DbName == dbName {
			m.flatItems[i].IsExpanded = !item.IsExpanded
			m.UpdateItems(m.dbs)
			return
		}
	}
}

func (m SidebarModel) View() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Height(m.height)

	content := TitleStyle.Bold(true).Render(" DATABASES") + "\n\n"

	if len(m.flatItems) == 0 {
		content += "  No databases found."
		return borderStyle.Width(m.width).Render(content)
	}

	maxVisible := m.height - 4
	startIdx := 0
	if m.cursor >= maxVisible {
		startIdx = m.cursor - maxVisible + 1
	}

	for i := startIdx; i < len(m.flatItems) && i-startIdx < maxVisible; i++ {
		item := m.flatItems[i]
		isSelected := i == m.cursor

		var line string
		if item.IsDb {
			icon := "▶ "
			if item.IsExpanded {
				icon = "▼ "
			}
			dbLabel := icon + item.DbName
			if isSelected {
				line = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorGreen)).
					Bold(true).
					Render(dbLabel)
			} else {
				line = DbStyle.Render(dbLabel)
			}
		} else {
			collLabel := "  " + item.CollName
			isActive := item.DbName == m.selectedDb && item.CollName == m.selectedColl
			if isSelected || isActive {
				line = CollSelectedStyle.Width(m.width - 4).Render(collLabel)
			} else {
				line = CollStyle.Render(collLabel)
			}
		}

		content += fmt.Sprintf("%s\n", line)
	}

	return borderStyle.Width(m.width).Render(content)
}
