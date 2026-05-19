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
)

type prepareModel struct {
	ctx *viewContext

	localDataSpinner   spinner.Model
	prepareTaskSpinner spinner.Model

	waitLocalData *promise.Await[[]ginger.SabunHash]
	waitPrepare   *promise.Await[[]candidateDownloadTask]

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
	localDataSpinner := spinner.New()
	localDataSpinner.Spinner = spinner.Dot
	prepareTaskSpinner := spinner.New()
	prepareTaskSpinner.Spinner = spinner.Dot

	// TODO: Download directory
	return prepareModel{
		ctx:                  ctx,
		localDataSpinner:     localDataSpinner,
		prepareTaskSpinner:   prepareTaskSpinner,
		waitLocalData:        promise.NewAwait[[]ginger.SabunHash](),
		waitPrepare:          promise.NewAwait[[]candidateDownloadTask](),
		ignoringMD5Hashes:    make(map[string]any),
		ignoringSHA256Hashes: make(map[string]any),
	}
}

func (m prepareModel) Init() tea.Cmd {
	go m.readLocalData(m.waitLocalData.Done)
	return tea.Batch(m.localDataSpinner.Tick, m.prepareTaskSpinner.Tick)
}

func (m prepareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	select {
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
		if !m.prepareDownloadTaskFlag && m.waitLocalData.Loaded {
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
	m.localDataSpinner, cmd = m.localDataSpinner.Update(msg)
	cmds = tea.Batch(cmds, cmd)
	m.prepareTaskSpinner, cmd = m.prepareTaskSpinner.Update(msg)
	cmds = tea.Batch(cmds, cmd)
	return m, cmds
}

func (m prepareModel) View() tea.View {
	s := ""
	if !m.prepareDownloadTaskFlag {
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

func (m prepareModel) readLocalData(done chan<- promise.Promise[[]ginger.SabunHash]) {
	data := make([]ginger.SabunHash, 0)
	conf := config.Snapshot.Load()
	switch conf.ClientType {
	case config.CLIENT_BEATORAJA:
		beatorajaScanner := bmsdb.NewBeatorajaScanner()
		scanResult, err := beatorajaScanner.ScanDirectory(conf.GameInstallationPath)
		if err != nil {
			done <- promise.Fail[[]ginger.SabunHash](err)
			return
		}
		reader := bmsdb.NewBeatorajaReader()
		songs, err := reader.SongData(bmsdb.NewQueryContext(scanResult.SongData))
		if err != nil {
			done <- promise.Fail[[]ginger.SabunHash](err)
			return
		}
		for _, song := range songs {
			data = append(data, ginger.SabunHash{
				MD5:    song.Md5,
				SHA256: song.Sha256,
			})
		}
	case config.CLIENT_LR2:
		lr2Scanner := bmsdb.NewLR2Scanner()
		scanResult, err := lr2Scanner.ScanDirectory(conf.GameInstallationPath)
		if err != nil {
			done <- promise.Fail[[]ginger.SabunHash](err)
			return
		}
		reader := bmsdb.NewLR2Reader()
		songs, err := reader.Song(bmsdb.NewQueryContext(scanResult.Song))
		if err != nil {
			done <- promise.Fail[[]ginger.SabunHash](err)
			return
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
		candidateTable := m.ctx.candidateTable
		for _, content := range candidateTable.Contents {
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
