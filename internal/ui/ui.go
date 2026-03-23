// Package ui: implementing the view of user interface
package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
	"github.com/Catizard/Ginger-Downloader/pkg/ginger"
)

const (
	SCENE_BOOT       = "BOOT"
	SCENE_WIZARD     = "WIZARD"
	SCENE_MENU       = "MENU"
	SCENE_TABLES     = "TABLES"
	SCENE_DOWNLOAD   = "DOWNLOAD"
	SCENE_SYNC_TABLE = "SYNC_TABLE"
	SCENE_PREPARE    = "PREPARE"
	SCENE_HELP       = "HELP"
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
	bootInfo        *bootInfo
	candidateHeader *ginger.TableHeader
	candidateTasks  []candidateDownloadTask
	conf            *config.Config
}

func InitMainController() MainController {
	ctx := viewContext{
		candidateTasks: make([]candidateDownloadTask, 0),
	}
	views := make(map[string]tea.Model)
	views[SCENE_BOOT] = InitializeBootModel(&ctx)
	views[SCENE_WIZARD] = InitializeWizardModel(&ctx)
	views[SCENE_MENU] = InitializeMenuModel()
	views[SCENE_TABLES] = InitializeTablesModel(&ctx)
	views[SCENE_DOWNLOAD] = InitializeDownloadModel(&ctx)
	views[SCENE_SYNC_TABLE] = InitializeSyncTableModel(&ctx)
	views[SCENE_PREPARE] = initializePrepareModel(&ctx)
	views[SCENE_HELP] = initializeHelpModel(&ctx)

	return MainController{
		current: views[SCENE_BOOT],
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
