package ui

import (
	"fmt"
	"log"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
	"github.com/Catizard/Ginger-Downloader/pkg/ginger"
)

type BootModel struct {
	s    spinner.Model
	done chan bootInfo
	ctx  *viewContext
}

type bootInfo struct {
	conf    *config.Config
	summary *ginger.ServerInfoSummary
	err     error
}

func newErrorBootInfo(err error) bootInfo {
	return bootInfo{
		conf:    nil,
		summary: nil,
		err:     err,
	}
}

func InitializeBootModel(ctx *viewContext) BootModel {
	done := make(chan bootInfo)
	go setupDownloader(done)
	return BootModel{
		s:    spinner.New(),
		done: done,
		ctx:  ctx,
	}
}

func (m BootModel) Init() tea.Cmd {
	return m.s.Tick
}

func (m BootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	select {
	case bootInfo := <-m.done:
		if bootInfo.err != nil {
			log.Fatalf("Failed to initialize downloader: %s", bootInfo.err)
		} else {
			m.ctx.bootInfo = &bootInfo
			m.ctx.conf = bootInfo.conf
			if bootInfo.conf.Initialized == 0 {
				return m, newTransferRequest(SCENE_WIZARD)
			} else {
				return m, newTransferRequest(SCENE_MENU)
			}
		}
	default:
		var cmd tea.Cmd
		m.s, cmd = m.s.Update(msg)
		return m, cmd
	}
	// How could this happen?
	return m, nil
}

func (m BootModel) View() tea.View {
	s := fmt.Sprintf("%s Initializing downloader...Please wait", m.s.View())
	return tea.NewView(s)
}

func setupDownloader(pushup chan<- bootInfo) {
	conf, err := config.ReadConfig()
	if err != nil {
		pushup <- newErrorBootInfo(err)
	}
	summary, err := ginger.Initialize()
	if err != nil {
		pushup <- newErrorBootInfo(err)
	}
	pushup <- bootInfo{
		conf,
		summary,
		nil,
	}
}
