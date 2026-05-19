// Package startup renders the startup errors
package startup

import (
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
)

type errorType int

const (
	NO_CONFIG_FILE = iota
	INVALID_CONFIG
	OTHERS
)

// startupErrorsModel is a standalone view that shows the errors
// during the startup phase.
type startupErrorsModel struct {
	err       error
	errorType errorType
}

func InitStartupErrorsModel(err error) *startupErrorsModel {
	return &startupErrorsModel{
		err:       err,
		errorType: distinguishError(err),
	}
}

func (model startupErrorsModel) Init() tea.Cmd {
	return nil
}

func (model startupErrorsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return model.handleKeyPresses(msg.String())
	}
	return model, nil
}

func (model startupErrorsModel) View() tea.View {
	switch model.errorType {
	case NO_CONFIG_FILE:
		help := `
No config file provided.
Do you want to create an empty config file?
Press E to create an english commented file;
C to create a chinese one
N to not create anything and exit
(Not case sensitive)`
		return tea.NewView(help)
	case INVALID_CONFIG:
		help := `
There's some error in your config file.
(Press any key to exit)`
		s := fmt.Sprintf("%s\n%s", help, model.err.Error())
		return tea.NewView(s)
	case OTHERS:
		help := `
There's some unexpected error happened.
If you don't know what to do, please contact us for support`
		s := fmt.Sprintf("%s\n%s", help, model.err.Error())
		return tea.NewView(s)
	}
	return tea.NewView("")
}

func (model startupErrorsModel) handleKeyPresses(kc string) (tea.Model, tea.Cmd) {
	switch model.errorType {
	case NO_CONFIG_FILE:
		localeChoosed := -1
		switch kc {
		case "n", "N":
			return model, tea.Quit
		case "e", "E":
			localeChoosed = config.ENGLISH_LOCALE
		case "c", "C":
			localeChoosed = config.CHINESE_LOCALE
		}

		if localeChoosed != -1 {
			if err := config.CreateTemplateConfig(config.Locale(localeChoosed), false); err != nil {
				model.err = err
				model.errorType = OTHERS
			} else {
				return model, tea.Quit
			}
		}
	case INVALID_CONFIG:
		return model, tea.Quit
	}
	return model, nil
}

func distinguishError(err error) errorType {
	if errors.Is(err, config.ErrorNoConfig) {
		return NO_CONFIG_FILE
	}
	var errInvalidConfig config.ErrorInvalidConfig
	if errors.As(err, &errInvalidConfig) {
		return INVALID_CONFIG
	}
	return OTHERS
}
