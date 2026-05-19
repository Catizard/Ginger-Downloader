package ui

import (
	"fmt"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
	"github.com/Catizard/Ginger-Downloader/internal/download"
)

var lock sync.Mutex

type DownloadModel struct {
	ctx             *viewContext
	downloadService *download.DownloadTaskService

	peekCount int

	taskCount    int
	waitCount    int
	runningCount int
	errCount     int

	tasks []*download.DownloadTask
}

func InitializeDownloadModel(ctx *viewContext) DownloadModel {
	// TODO: Download directory
	return DownloadModel{
		ctx:       ctx,
		peekCount: 5,
	}
}

func (m DownloadModel) Init() tea.Cmd {
	return tickEvery(1 * time.Second)
}

func (m DownloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	lock.Lock()
	if m.downloadService == nil {
		conf := config.Snapshot.Load()
		downloadDirectory := conf.DownloadDirectory
		m.downloadService = download.NewDownloadTaskService(downloadDirectory, 5)
		go m.submitDownloadTasks()
	}
	lock.Unlock()
	switch msg.(type) {
	case tickMsg:
		if m.downloadService != nil {
			m.taskCount, m.waitCount, m.runningCount, m.errCount = m.downloadService.InternalTaskCount()
			m.tasks = m.downloadService.PeekRunningTasks(m.peekCount)
		}
		return m, tickEvery(1 * time.Second)
	}
	return m, nil
}

func (m DownloadModel) View() tea.View {
	s := ""
	s += fmt.Sprintf("total: %d, waiting: %d, running: %d, error: %d\n", m.taskCount, m.waitCount, m.runningCount, m.errCount)
	s += "Press Ctrl+C halts the download process\n"
	return tea.NewView(s)
}

func (m DownloadModel) submitDownloadTasks() {
	candidateTasks := m.ctx.candidateTasks
	for _, task := range candidateTasks {
		if task.UseMD5 {
			m.downloadService.SubmitSingleMD5DownloadTask(task.MD5, &task.Name)
		} else {
			// TODO: Sha256
		}
	}
}

type tickMsg struct{}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
