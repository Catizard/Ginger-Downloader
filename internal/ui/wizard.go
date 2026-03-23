package ui

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
)

var confirmed bool

type WizardModel struct {
	form     *huh.Form
	formData *formData
	errorMsg string

	ctx *viewContext
}

type formData struct {
	type_             string
	database          string
	downloadDirectory string
}

func InitializeWizardModel(ctx *viewContext) WizardModel {
	formData := &formData{}
	return WizardModel{
		formData: formData,
		form:     newForm(formData),
		ctx:      ctx,
	}
}

func (m WizardModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		if tryConnectingLocalDatabase() {
			conf := m.ctx.conf
			conf.ClientType = config.ClientType(m.formData.type_)
			conf.Initialized = 1
			conf.LocalDBPath = m.formData.database
			conf.DownloadDirectory = m.formData.downloadDirectory
			if err := conf.WriteConfig(); err != nil {
				log.Fatalf("Failed to write config: %s", err)
			}
			m.ctx.conf = conf
			return m, newTransferRequest(SCENE_MENU)
		} else {
			m.form = newForm(m.formData)
			m.errorMsg = "cannot connect to local database, please ensure you're picking a correct file"
			return m, m.form.Init()
		}
	}
	return m, cmd
}

func (m WizardModel) View() tea.View {
	s := "Hey, we need to set some basic info before you start downloading things"
	s += "\n"
	if m.errorMsg != "" {
		s += m.errorMsg + "\n"
	}
	s += m.form.View()
	return tea.NewView(s)
}

func newForm(data *formData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("type").
				Options(huh.NewOptions("Beatoraja", "LR2")...).
				Title("What client do you use to play?").
				Value(&data.type_),

			huh.NewFilePicker().
				AllowedTypes([]string{".db"}).
				Key("database").
				CurrentDirectory("/").
				Title("Choose your database file").
				Value(&data.database),

			huh.NewFilePicker().
				Key("downloadDirectory").
				CurrentDirectory("/").
				FileAllowed(false).
				DirAllowed(true).
				Title("Choose where you want download to").
				Value(&data.downloadDirectory),
		),
	)
}

func tryConnectingLocalDatabase() bool {
	return true
}
