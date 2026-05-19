package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
)

// InitializeLog initializes the log file
func InitializeLog() func() {
	f, err := tea.LogToFile("downloader.debug", "debug")
	if err != nil {
		log.Fatal(err)
	}
	f.Truncate(0)
	f.Seek(0, 0)

	log.SetOutput(f)

	log.Printf("Hey! Welcome to Ginger's downloader. Current build's version is %s\n", config.VERSION)
	log.Printf("This config file might contain your sensitive data. E.g. your actual name in the directory path\n")
	log.Printf("Before you share this file for any purpose, check whether it has or not!\n")
	log.Printf("If you don't care about it then leave it as is :)\n")

	return func() {
		f.Close()
	}
}
