package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Catizard/Ginger-Downloader/pkg/ginger"
	"github.com/imroc/req/v3"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type DownloadTaskService struct {
	mutex                sync.Mutex
	tasks                []*DownloadTask
	waitTasks            []*DownloadTask
	runningTasks         map[uint]*DownloadTask
	updMsgReceiver       chan taskUpdMsg
	maximumDownloadCount int
	downloadDirectory    string
	// Exeprimental, only for test case
	taskID   uint
	errCount int
}

type taskUpdMsg struct {
	taskID        uint
	final         bool
	err           error
	downloadSize  int64
	contentLength int64
}

func NewDownloadTaskService(downloadDirectory string, maximumDownloadCount int) *DownloadTaskService {
	service := &DownloadTaskService{
		tasks:                make([]*DownloadTask, 0),
		waitTasks:            make([]*DownloadTask, 0),
		runningTasks:         make(map[uint]*DownloadTask),
		taskID:               1,
		updMsgReceiver:       make(chan taskUpdMsg),
		downloadDirectory:    downloadDirectory,
		maximumDownloadCount: maximumDownloadCount,
	}
	go service.receive()
	return service
}

func (s *DownloadTaskService) lock() {
	s.mutex.Lock()
}

func (s *DownloadTaskService) unlock() {
	s.mutex.Unlock()
}

// For internal test, return: task count, wait task count, running task count
func (s *DownloadTaskService) InternalTaskCount() (int, int, int, int) {
	s.lock()
	defer s.unlock()
	return len(s.tasks), len(s.waitTasks), len(s.runningTasks), s.errCount
}

func (s *DownloadTaskService) PeekRunningTasks(count int) []*DownloadTask {
	s.lock()
	defer s.unlock()
	ret := make([]*DownloadTask, 0)
	// TODO: This shit is not stable
	for _, task := range s.runningTasks {
		ret = append(ret, task)
		if len(ret) == count {
			break
		}
	}
	return ret
}

// DownloadTaskService's life cycle, receive other routine's message
// and update internal states
func (s *DownloadTaskService) receive() {
	for {
		select {
		case msg := <-s.updMsgReceiver:
			s.handleUpdateTask(&msg)
		default:
		}
		s.tryKickingWaitTask()
		time.Sleep(150 * time.Millisecond)
	}
}

func (s *DownloadTaskService) handleUpdateTask(msg *taskUpdMsg) error {
	s.lock()
	defer s.unlock()
	if task, ok := s.runningTasks[msg.taskID]; ok {
		if msg.final {
			delete(s.runningTasks, msg.taskID)
			if msg.err != nil {
				log.Printf("[DownloadTaskService] task %d fails: %s", msg.taskID, msg.err)
				*task.Status = TASK_ERROR
				task.DownloadSize = 0
				task.ContentLength = 0
				task.ErrorMessage = msg.err.Error()
				s.errCount++
			} else {
				log.Printf("[DownloadTaskService] task %d done", msg.taskID)
				*task.Status = TASK_SUCCESS
			}
		} else {
			*task.Status = TASK_DOWNLOAD
			task.DownloadSize = msg.downloadSize
			task.ContentLength = msg.contentLength
		}
	} else {
		log.Printf("[DownloadTaskService] discard updte msg: %v", *msg)
	}
	return nil
}

func (s *DownloadTaskService) submitTaskError(taskID uint, err error) {
	s.updMsgReceiver <- taskUpdMsg{
		taskID: taskID,
		final:  true,
		err:    err,
	}
}

func (s *DownloadTaskService) tryKickingWaitTask() {
	s.lock()
	defer s.unlock()
	if s.maximumDownloadCount == len(s.runningTasks) || len(s.waitTasks) == 0 {
		return
	}
	next := s.waitTasks[0]
	taskID := next.ID
	log.Printf("[DownloadTaskService] try kicking task %d(%s)", taskID, next.URL)
	s.waitTasks = s.waitTasks[1:]
	s.runningTasks[taskID] = next
	go func() {
		// Open a no timeout, cancelable client
		client := req.C().SetTimeout(0).SetCommonRetryCount(-1)
		// Prevent a very rare race condition?
		s.lock()
		ctx, cancel := context.WithCancel(context.Background())
		s.runningTasks[taskID].Cancel = cancel
		contextLockedReq := client.R().SetContext(ctx)
		s.unlock()
		resp, err := contextLockedReq.
			SetOutputFile(next.IntermediateFilePath).
			SetDownloadCallbackWithInterval(func(info req.DownloadInfo) {
				if info.Response != nil && info.Response.Response != nil {
					contentLength := info.Response.Response.ContentLength
					s.updMsgReceiver <- taskUpdMsg{
						taskID:        next.ID,
						final:         false,
						err:           nil,
						downloadSize:  info.DownloadedSize,
						contentLength: contentLength,
					}
				} else {
					log.Printf("invalid http download response, what is happening?")
				}
			}, 1*time.Second).
			Get(next.URL)
		if err != nil {
			s.submitTaskError(next.ID, err)
			return
		}
		if !resp.IsSuccessState() {
			// NOTE: At this moment, the content should be placed at file
			// Lampghost will try read the error response data, if anything went wrong,
			// just cancel it
			f, err := os.Open(next.IntermediateFilePath)
			if err != nil {
				s.submitTaskError(next.ID, eris.New("remote server returns an unexpected error"))
				return
			}
			body, err := io.ReadAll(f)
			if err != nil {
				s.submitTaskError(next.ID, eris.New("remote server returns an unexpected error"))
				return
			}
			s.submitTaskError(next.ID, errors.New(string(body)))
			return
		}
		filename := ""
		if next.TaskName != nil {
			filename = *next.TaskName
		}
		if next.FallbackName != "" {
			filename = next.FallbackName
		}
		contentDisposition := resp.GetHeader("Content-Disposition")
		log.Printf("Content-Disposition: %s", contentDisposition)
		if contentDisposition != "" {
			if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
				filename = params["filename"]
			} else {
				log.Printf("[DownloadTaskService] cannot parse media type from Content-Disposition")
			}
		} else {
			log.Printf("[DownloadTaskService] cannot fetch Content-Disposition from response")
		}
		// NOTE: <del>Below check & conversion was stolen from wriggle, sorry wriggle!</del>
		if filename == "" || filename == "/" || filename == "." {
			s.submitTaskError(next.ID, eris.New("cannot determine filename"))
			return
		}

		filename = filepath.Clean(filename)
		filename = strings.ReplaceAll(filename, "/", "_")
		filename = strings.ReplaceAll(filename, "\\", "_")
		filename = strings.ReplaceAll(filename, ":", "_")
		filename = strings.ReplaceAll(filename, "*", "_")
		filename = strings.ReplaceAll(filename, "?", "_")
		filename = strings.ReplaceAll(filename, "\"", "_")
		filename = strings.ReplaceAll(filename, "<", "_")
		filename = strings.ReplaceAll(filename, ">", "_")
		filename = strings.ReplaceAll(filename, "|", "_")

		targetPath := filepath.Join(s.downloadDirectory, filename)

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			s.submitTaskError(next.ID, eris.Wrapf(err, "cannot create directory for %s", targetPath))
			return
		}
		if _, err := os.Stat(targetPath); err == nil {
			log.Printf("[DownloadTaskService] target file is already existed, would be replaced with the current one")
		} else if !os.IsNotExist(err) {
			s.submitTaskError(next.ID, eris.Errorf("unexpected stat(%s) error: %s", targetPath, err))
			return
		}
		if err := os.Rename(next.IntermediateFilePath, targetPath); err != nil {
			s.submitTaskError(next.ID, eris.Wrap(err, "cannot rename"))
			return
		}

		// Everything is done
		s.updMsgReceiver <- taskUpdMsg{
			taskID: next.ID,
			final:  true,
			err:    nil,
		}
	}()
}

func (s *DownloadTaskService) SubmitSingleMD5DownloadTask(md5 string, taskName *string) error {
	if s.downloadDirectory == "" {
		return eris.New("download directory cannot be empty")
	}
	if md5 == "" {
		return eris.New("assert: md5 cannot be empty")
	}
	downloadInfo, err := ginger.GingerDownloadSource.GetDownloadURLFromMD5(md5)
	if err != nil {
		return eris.Wrap(err, "build download url")
	}
	log.Printf("unique symbol: %s", downloadInfo.UniqueSymbol)
	s.lock()
	if downloadInfo.UniqueSymbol == "" {
		log.Printf("download task's unique symbol is empty string")
	} else {
		for _, task := range s.tasks {
			if task.UniqueSymbol == downloadInfo.UniqueSymbol {
				log.Printf("skipping download task due to have same unique symbol: %s", task.UniqueSymbol)
				s.unlock()
				return eris.New("duplicated download task")
			}
		}
	}
	s.unlock()
	log.Printf("[DownloadTaskService] build url: %s", downloadInfo.DownloadURL)
	currentTaskID := s.taskID
	s.taskID++
	intermediateFileName := fmt.Sprintf("%d.crdownload", currentTaskID)
	return s.submitSingleDownloadTask(currentTaskID, downloadInfo, intermediateFileName, taskName)
}

func (s *DownloadTaskService) submitSingleDownloadTask(id uint, downloadInfo ginger.DownloadInfo, intermediateFileName string, taskName *string) error {
	if s.downloadDirectory == "" {
		return eris.New("download directory cannot be empty")
	}
	intermediateFilePath := filepath.Join(s.downloadDirectory, intermediateFileName)
	s.lock()
	defer s.unlock()
	status := TASK_PREPARE
	task := DownloadTask{
		Model: gorm.Model{
			ID: id,
		},
		URL:                  downloadInfo.DownloadURL,
		Status:               &status,
		IntermediateFilePath: intermediateFilePath,
		FallbackName:         downloadInfo.FileName,
		TaskName:             taskName,
		DownloadSize:         0,
		ContentLength:        0,
		UniqueSymbol:         downloadInfo.UniqueSymbol,
	}
	s.tasks = append(s.tasks, &task)
	s.waitTasks = append(s.waitTasks, &task)
	return nil
}

// Query a current snapshot of download tasks
func (s *DownloadTaskService) FindDownloadTaskList() ([]*DownloadTask, int, error) {
	s.lock()
	defer s.unlock()
	ret := make([]*DownloadTask, len(s.tasks))
	copy(ret, s.tasks)
	slices.Reverse(ret)
	return ret, len(ret), nil
}

func (s *DownloadTaskService) CancelDownloadTask(taskID uint) error {
	s.lock()
	defer s.unlock()
	if task, ok := s.runningTasks[taskID]; ok {
		if task.Cancel != nil {
			*task.Status = TASK_CANCEL
			task.Cancel()
			delete(s.runningTasks, taskID)
			// NOTE: Now, the canceld task is not in wait queue nor running queue
			// It's only referenced in all task list, requring a 'Restart' command
			// to rejoin the party
		}
	}
	return nil
}

func (s *DownloadTaskService) RestartDownloadTask(taskID uint) error {
	s.lock()
	defer s.unlock()
	// Won't be a problem for now
	for _, task := range s.tasks {
		if task.ID == taskID {
			if *task.Status != TASK_CANCEL && *task.Status != TASK_ERROR {
				return eris.New("assert: cannot restart a task is not canceled or failed")
			}
			*task.Status = TASK_PREPARE
			s.waitTasks = append(s.waitTasks, task)
			break
		}
	}
	return nil
}
