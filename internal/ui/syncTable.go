package ui

import tea "charm.land/bubbletea/v2"

type syncTableModel struct {
	ctx *viewContext
}

func InitializeSyncTableModel(ctx *viewContext) syncTableModel {
	return syncTableModel{
		ctx: ctx,
	}
}

func (m syncTableModel) Init() tea.Cmd {
	return nil
}

func (m syncTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m syncTableModel) View() tea.View {
	return tea.NewView("Sync table view")
}
