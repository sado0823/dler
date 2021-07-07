package downloader

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal"
	"github.com/sado0823/downloader/internal/core"
	"github.com/sado0823/downloader/internal/tool"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var conNum int64

func init() {
	RootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().Int64VarP(&conNum, "goroutines count", "c", int64(runtime.NumCPU()), "default is your CPU threads count")

}

var downloadCmd = &cobra.Command{
	Use:     "download",
	Short:   "downloads a file from URL or file name",
	Example: `downloader [-c=goroutines_count] URL`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		core.ExitWithErr(download(args))
	},
}

func download(args []string) error {
	folder, err := tool.DownloaderFolder(args[0])
	if err != nil {
		return errors.WithStack(err)
	}
	if tool.FolderExist(folder) {
		fmt.Printf("Task already exist, remove it first \n")
		folder, err = tool.DownloaderFolder(args[0])
		if err != nil {
			return errors.WithStack(err)
		}
		if err := os.RemoveAll(folder); err != nil {
			return errors.WithStack(err)
		}
	}

	return internal.Do(args[0], nil, conNum)
}
