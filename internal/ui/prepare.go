package ui

import (
	"fmt"
	"log"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/Catizard/Ginger-Downloader/internal/config"
	"github.com/Catizard/Ginger-Downloader/internal/promise"
	"github.com/Catizard/Ginger-Downloader/pkg/ginger"
	"github.com/Catizard/bmsdb"
	"github.com/Catizard/bmstable"
)

type prepareModel struct {
	ctx *viewContext

	downloadSpinner    spinner.Model
	localDataSpinner   spinner.Model
	prepareTaskSpinner spinner.Model

	waitTable     *promise.Await[bmstable.DifficultTable]
	waitLocalData *promise.Await[[]ginger.SabunHash]
	waitPrepare   *promise.Await[[]candidateDownloadTask]

	candidateTable          bmstable.DifficultTable
	ignoringMD5Hashes       map[string]any
	ignoringSHA256Hashes    map[string]any
	prepareDownloadTaskFlag bool
}

type candidateDownloadTask struct {
	Name   string
	MD5    string
	SHA256 string
	UseMD5 bool
}

func initializePrepareModel(ctx *viewContext) prepareModel {
	downloadSpinner := spinner.New()
	downloadSpinner.Spinner = spinner.Dot
	localDataSpinner := spinner.New()
	localDataSpinner.Spinner = spinner.Dot
	prepareTaskSpinner := spinner.New()
	prepareTaskSpinner.Spinner = spinner.Dot

	// TODO: Download directory
	return prepareModel{
		ctx:                  ctx,
		downloadSpinner:      downloadSpinner,
		localDataSpinner:     localDataSpinner,
		prepareTaskSpinner:   prepareTaskSpinner,
		waitTable:            promise.NewAwait[bmstable.DifficultTable](),
		waitLocalData:        promise.NewAwait[[]ginger.SabunHash](),
		waitPrepare:          promise.NewAwait[[]candidateDownloadTask](),
		ignoringMD5Hashes:    make(map[string]any),
		ignoringSHA256Hashes: make(map[string]any),
	}
}

func (m prepareModel) Init() tea.Cmd {
	go m.downloadTableData(m.ctx.candidateHeader.HeaderURL, m.waitTable.Done)
	go m.readLocalData(m.waitLocalData.Done)
	return tea.Batch(m.downloadSpinner.Tick, m.localDataSpinner.Tick, m.prepareTaskSpinner.Tick)
}

func (m prepareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	select {
	case p := <-m.waitTable.Done:
		if p.Err != nil {
			// TODO: Do something here
			log.Fatalf("something fatal happened: %s", p.Err)
		}
		m.candidateTable = *p.Data
		m.waitTable.Loaded = true
		if !m.prepareDownloadTaskFlag && m.waitLocalData.Loaded && m.waitTable.Loaded {
			go m.prepareTasksToDownload(m.waitPrepare.Done)
			m.prepareDownloadTaskFlag = true
		}
	case p := <-m.waitLocalData.Done:
		if p.Err != nil {
			log.Fatalf("something fatal happened: %s", p.Err)
		}
		rawData := *p.Data
		for _, hash := range rawData {
			if hash.MD5 != "" {
				m.ignoringMD5Hashes[hash.MD5] = new(any)
			}
			if hash.SHA256 != "" {
				m.ignoringSHA256Hashes[hash.SHA256] = new(any)
			}
		}
		m.waitLocalData.Loaded = true
		if !m.prepareDownloadTaskFlag && m.waitLocalData.Loaded && m.waitTable.Loaded {
			go m.prepareTasksToDownload(m.waitPrepare.Done)
			m.prepareDownloadTaskFlag = true
		}
	case p := <-m.waitPrepare.Done:
		if p.Err != nil {
			log.Fatalf("something fatal happened: %s", p.Err)
		}
		m.ctx.candidateTasks = *p.Data
		m.waitPrepare.Loaded = true
		return m, newTransferRequest(SCENE_DOWNLOAD)
	default:
	}

	var cmd tea.Cmd = nil
	var cmds tea.Cmd
	m.downloadSpinner, cmd = m.downloadSpinner.Update(msg)
	cmds = tea.Batch(cmds, cmd)
	m.localDataSpinner, cmd = m.localDataSpinner.Update(msg)
	cmds = tea.Batch(cmds, cmd)
	m.prepareTaskSpinner, cmd = m.prepareTaskSpinner.Update(msg)
	cmds = tea.Batch(cmds, cmd)
	return m, cmds
}

func (m prepareModel) View() tea.View {
	s := ""
	if !m.prepareDownloadTaskFlag {
		if !m.waitTable.Loaded {
			s += fmt.Sprintf("%s Downloading table data\n", m.downloadSpinner.View())
		} else {
			s += "Table data downloaded\n"
		}
		if !m.waitLocalData.Loaded {
			s += fmt.Sprintf("%s Reading local data\n", m.localDataSpinner.View())
		} else {
			s += "Local data fetched"
		}
	} else {
		if !m.waitPrepare.Loaded {
			s += fmt.Sprintf("%s Preparing download tasks\n", m.prepareTaskSpinner.View())
		} else {
			s += "Download task prepared"
		}
	}

	return tea.NewView(s)
}

func (m prepareModel) downloadTableData(tableURL string, done chan<- promise.Promise[bmstable.DifficultTable]) {
	dt, err := bmstable.ParseFromURL(tableURL)
	if err != nil {
		done <- promise.Fail[bmstable.DifficultTable](err)
		return
	}

	done <- promise.Ok(&dt)
}

func (m prepareModel) readLocalData(done chan<- promise.Promise[[]ginger.SabunHash]) {
	data := make([]ginger.SabunHash, 0)
	conf := m.ctx.conf
	switch conf.ClientType {
	case config.CLIENT_BEATORAJA:
		reader := bmsdb.NewBeatorajaReader()
		songs, err := reader.SongData(bmsdb.NewQueryContext(conf.LocalDBPath))
		if err != nil {
			done <- promise.Fail[[]ginger.SabunHash](err)
		}
		for _, song := range songs {
			data = append(data, ginger.SabunHash{
				MD5:    song.Md5,
				SHA256: song.Sha256,
			})
		}
	case config.CLIENT_LR2:
		reader := bmsdb.NewLR2Reader()
		songs, err := reader.Song(bmsdb.NewQueryContext(conf.LocalDBPath))
		if err != nil {
			done <- promise.Fail[[]ginger.SabunHash](err)
		}
		for _, song := range songs {
			data = append(data, ginger.SabunHash{
				MD5: song.MD5,
			})
		}
	}
	log.Printf("local data count: %d", len(data))
	done <- promise.Ok(&data)
}

func (m prepareModel) prepareTasksToDownload(done chan<- promise.Promise[[]candidateDownloadTask]) {
	go func() {
		data := make([]candidateDownloadTask, 0)
		for _, content := range m.candidateTable.Contents {
			candidate := candidateDownloadTask{
				Name: content.Title,
			}
			if content.Md5 != "" {
				if _, ok := m.ignoringMD5Hashes[content.Md5]; ok {
					continue
				}
				candidate.UseMD5 = true
				candidate.MD5 = content.Md5
			} else if content.Sha256 != "" {
				if _, ok := m.ignoringSHA256Hashes[content.Sha256]; ok {
					continue
				}
				candidate.UseMD5 = false
				candidate.SHA256 = content.Sha256
			}
			data = append(data, candidate)
		}
		done <- promise.Ok(&data)
	}()
}
