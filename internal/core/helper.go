package core

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal/tool"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func TaskPrint() error {

	fileInfos, err := ioutil.ReadDir(tool.DownloaderHome())
	if err != nil {
		return errors.WithStack(err)
	}

	folders := make([]string, 0)
	for _, info := range fileInfos {
		if info.IsDir() {
			folders = append(folders, info.Name())
		}
	}

	folderStr := strings.Join(folders, "\n")
	fmt.Printf("Currently on going download: \n")
	fmt.Println(folderStr)

	return nil
}

func ErrPrint(err error) {
	if err != nil {
		log.Printf("%+v\n", errors.Cause(err))
		os.Exit(1)
	}
}
