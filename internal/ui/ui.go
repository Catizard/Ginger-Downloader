// Package ui: implementing the view of user interface
package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/bmstable"
)

const (
	SCENE_TABLES   = "TABLES"
	SCENE_DOWNLOAD = "DOWNLOAD"
	SCENE_PREPARE  = "PREPARE"
)

type transferMsg string

func newTransferRequest(next string) tea.Cmd {
	return func() tea.Msg {
		return transferMsg(next)
	}
}

// MainController is the facade of the whole view
type MainController struct {
	current tea.Model
	views   map[string]tea.Model
	ctx     *viewContext
}

// ViewContext is the 'global' state shared for all views to read and write
type viewContext struct {
	candidateTasks []candidateDownloadTask
	candidateTable bmstable.DifficultTable
}

func InitMainController() MainController {
	ctx := viewContext{
		candidateTasks: make([]candidateDownloadTask, 0),
	}
	views := make(map[string]tea.Model)
	views[SCENE_TABLES] = InitializeTablesModel(&ctx)
	views[SCENE_DOWNLOAD] = InitializeDownloadModel(&ctx)
	views[SCENE_PREPARE] = initializePrepareModel(&ctx)

	return MainController{
		current: views[SCENE_TABLES],
		views:   views,
		ctx:     &ctx,
	}
}

func (m MainController) Init() tea.Cmd {
	return m.current.Init()
}

func (m MainController) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case transferMsg:
		m.current = m.views[string(msg)]
		return m, m.current.Init()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			m.current, cmd = m.current.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.current, cmd = m.current.Update(msg)
		return m, cmd
	}
}

func (m MainController) View() tea.View {
	return m.current.View()
}
