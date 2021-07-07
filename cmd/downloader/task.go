package downloader

import (
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(taskCmd)
}

var taskCmd = &cobra.Command{
	Use:     "task",
	Short:   "show current downloading task",
	Example: `downloader task`,
	Run: func(cmd *cobra.Command, args []string) {
		core.ExitWithErr(task())
	},
}

func task() error {
	err := core.TaskPrint()
	return errors.WithStack(err)
}
