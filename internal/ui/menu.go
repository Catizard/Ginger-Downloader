package ui

import tea "charm.land/bubbletea/v2"

type MenuModel struct {
	cursor  int
	options []menuOption
}

type menuOption struct {
	title string
	hook  func(m MenuModel) (MenuModel, tea.Cmd)
}

func InitializeMenuModel() MenuModel {
	return MenuModel{
		cursor: 0,
		options: []menuOption{
			{
				"Select a table and start downloading",
				handleDownloadBasedOnTable,
			},
			{
				"Display current configurations",
				handleDisplayCurrentConfigurations,
			},
			{
				"Modify configurations",
				handleModifyConfigurations,
			},
			{
				"Help",
				handleHelp,
			},
			{
				"Leave",
				handleLeave,
			},
		},
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(m.options)) % len(m.options)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.options)
		case "enter":
			return m.options[m.cursor].hook(m)
		}
	}
	return m, nil
}

func (m MenuModel) View() tea.View {
	s := "Hey! What do you want to do today?\n"
	for i, option := range m.options {
		if i == m.cursor {
			s += "> "
		} else {
			s += "  "
		}
		s += option.title + "\n"
	}
	return tea.NewView(s)
}

func handleDownloadBasedOnTable(m MenuModel) (MenuModel, tea.Cmd) {
	return m, newTransferRequest(SCENE_TABLES)
}

func handleDisplayCurrentConfigurations(m MenuModel) (MenuModel, tea.Cmd) {
	return m, nil
}

func handleModifyConfigurations(m MenuModel) (MenuModel, tea.Cmd) {
	return m, nil
}

func handleHelp(m MenuModel) (MenuModel, tea.Cmd) {
	return m, newTransferRequest(SCENE_HELP)
}

func handleLeave(m MenuModel) (MenuModel, tea.Cmd) {
	return m, tea.Quit
}
