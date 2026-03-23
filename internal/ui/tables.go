package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type TablesModel struct {
	cursor int
	err    error
	ctx    *viewContext
}

func InitializeTablesModel(ctx *viewContext) TablesModel {
	return TablesModel{
		ctx: ctx,
	}
}

func (m TablesModel) Init() tea.Cmd {
	return nil
}

func (m TablesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ctx.bootInfo != nil {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "up", "k":
				m.cursor = (m.cursor - 1 + m.lenHeaders()) % m.lenHeaders()
			case "down", "j":
				m.cursor = (m.cursor + 1) % m.lenHeaders()
			case "enter":
				header := m.ctx.bootInfo.summary.Headers[m.cursor]
				m.ctx.candidateHeader = &header
				return m, newTransferRequest(SCENE_PREPARE)
			case "esc":
				return m, newTransferRequest(SCENE_MENU)
			}
		}
	}
	return m, nil
}

func (m TablesModel) View() tea.View {
	if m.ctx.bootInfo == nil {
		return tea.NewView("No boot info passed, please restart the program to try again or fire a new bug report!")
	}

	s := ""
	for i, header := range m.ctx.bootInfo.summary.Headers {
		if i == m.cursor {
			s += "> "
		} else {
			s += "  "
		}
		s += fmt.Sprintf("%s (%d/%d)\n", header.Name, header.DataCount-header.MissingCount, header.DataCount)
	}

	return tea.NewView(s)
}

func (m TablesModel) lenHeaders() int {
	if m.ctx.bootInfo == nil {
		return 0
	}
	return len(m.ctx.bootInfo.summary.Headers)
}
