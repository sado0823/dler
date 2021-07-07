package downloader


import (
	"os"

	"github.com/spf13/cobra"
)

// when cli called without any child commands
var RootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "file downloader written in Go",
}