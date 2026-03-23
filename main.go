package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.InitMainController())
	f, err := tea.LogToFile("downloader.debug", "debug")
	if err != nil {
		log.Fatal(err)
	}
	f.Truncate(0)
	f.Seek(0, 0)
	defer f.Close()
	log.SetOutput(f)
	if _, err := p.Run(); err != nil {
		os.Exit(-1)
	}
}
