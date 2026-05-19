package main

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
	"github.com/Catizard/Ginger-Downloader/internal/startup"
	"github.com/Catizard/Ginger-Downloader/internal/ui"
)

func main() {
	cleanup := InitializeLog()
	defer cleanup()

	_, err := config.InitConfig()
	if err != nil {
		p := tea.NewProgram(startup.InitStartupErrorsModel(err))
		if _, err := p.Run(); err != nil {
			os.Exit(-1)
		}
	}

	p := tea.NewProgram(ui.InitMainController())
	if _, err := p.Run(); err != nil {
		os.Exit(-1)
	}
}
