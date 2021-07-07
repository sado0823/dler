package internal

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal/core"
	"github.com/sado0823/downloader/internal/tool"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func Do(url string, state *core.State, conNum int64) error {

	// signal handle
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,  // 终端挂起或者控制进程终止
		syscall.SIGINT,  // 键盘中断
		syscall.SIGTERM, // 终止信号
		syscall.SIGQUIT, // 键盘的退出键被按下
	)

	files := make([]string, 0)
	ranges := make([]core.DownloadRange, 0)
	isInterrupted := false

	doneChan := make(chan bool, conNum)
	fileChan := make(chan string, conNum)
	errorChan := make(chan error, 1)
	stateChan := make(chan core.DownloadRange, 1)
	interruptChan := make(chan bool, conNum)

	var downloader *core.Downloader

	var err error

	if state == nil {
		downloader, err = core.NewDownloader(url, conNum)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		downloader = &core.Downloader{
			Url:            state.URL,
			FileName:       filepath.Base(state.URL),
			Part:           int64(len(state.Range)),
			SkipTls:        true,
			DownloadRanges: state.Range,
			Reusable:       true,
		}
	}

	// start downloading
	go downloader.Downloading(doneChan, fileChan, errorChan, interruptChan, stateChan)

	// monitoring the download in progress
	for {
		select {
		case <-signalChan:
			isInterrupted = true
			for conNum > 0 {
				interruptChan <- true
				conNum--
			}
		case file := <-fileChan:
			files = append(files, file)
		case err = <-errorChan:
			return errors.WithStack(err)
		case part := <-stateChan:
			ranges = append(ranges, part)
		case <-doneChan:
			if isInterrupted {
				if downloader.Reusable {
					fmt.Printf("Interrupted, saving state ... \n")
					s := &core.State{
						URL:   url,
						Range: ranges,
					}
					if err = s.Save(); err != nil {
						return errors.WithStack(err)
					}
					return nil
				} else {
					fmt.Printf("Interrupted, but downloading url is not resumable, silently die\t\n")
					return nil
				}
			} else {
				err = tool.MergeFile(files, filepath.Base(url))
				if err != nil {
					return errors.WithStack(err)
				}

				folder, err := tool.DownloaderFolder(url)
				if err != nil {
					return errors.WithStack(err)
				}

				err = os.RemoveAll(folder)
				if err != nil {
					return errors.WithStack(err)
				}

				return nil
			}
		}
	}
}
