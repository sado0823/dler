package downloader

import (
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal"
	"github.com/sado0823/downloader/internal/core"
	"github.com/sado0823/downloader/internal/tool"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(resumeCmd)
}

var resumeCmd = &cobra.Command{
	Use:     "resume",
	Short:   "resume downloading task",
	Example: `downloader resume URL or file name`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		core.ExitWithErr(resumeTask(args))
	},
}

func resumeTask(args []string) error {
	task := ""
	if !tool.InvalidURL(args[0]) {
		task = tool.DownloaderName(args[0])
	} else {
		task = args[0]
	}

	state, err := core.Resume(task)
	if err != nil {
		return errors.WithStack(err)
	}
	return internal.Do(state.URL, state, conNum)
}
