package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
	"github.com/Catizard/Ginger-Downloader/internal/download"
)

var lock sync.Mutex

type DownloadModel struct {
	ctx             *viewContext
	downloadService *download.DownloadTaskService
	viewport        viewport.Model

	initialziedViewport bool

	peekCount int

	taskCount    int
	waitCount    int
	runningCount int
	errCount     int

	tasks []*download.DownloadTask
	logs  []string
}

func InitializeDownloadModel(ctx *viewContext) DownloadModel {
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
	var cmds tea.Cmd
	if m.downloadService == nil {
		conf := config.Snapshot.Load()
		downloadDirectory := conf.DownloadDirectory
		m.downloadService = download.NewDownloadTaskService(downloadDirectory, 5)
		go m.submitDownloadTasks()
		cmds = tea.Batch(cmds, tea.RequestWindowSize)
	}
	lock.Unlock()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.initialziedViewport {
			m.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height-2))
			m.viewport.YPosition = 0
			m.initialziedViewport = true
		}
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - 2)
	case tickMsg:
		if m.downloadService != nil {
			m.taskCount, m.waitCount, m.runningCount, m.errCount = m.downloadService.InternalTaskCount()
			m.tasks = m.downloadService.PeekRunningTasks(m.peekCount)
			m.logs = m.downloadService.GlanceLogs()
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.viewport.GotoBottom()
		}
		cmds = tea.Batch(cmds, tickEvery(1*time.Second))
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, tea.Batch(cmds, cmd)
}

func (m DownloadModel) View() tea.View {
	s := ""
	s += fmt.Sprintf("total: %d, waiting: %d, running: %d, error: %d\n", m.taskCount, m.waitCount, m.runningCount, m.errCount)

	s += "Press Ctrl+C halts the download process\n"
	if !m.initialziedViewport {
		s += "Initializing\n"
	} else {
		s += m.viewport.View()
	}
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
