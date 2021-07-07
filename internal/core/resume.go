package core

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal/consts"
	"github.com/sado0823/downloader/internal/tool"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	DownloadRange struct {
		URL  string
		Path string
		From int64
		To   int64
	}
	State struct {
		URL   string
		Range []DownloadRange
	}
)

func (s *State) Save() error {

	// save dir
	folder, err := tool.DownloaderFolder(s.URL)
	if err != nil {
		return err
	}

	err = tool.Mkdir(folder)
	if err != nil {
		return err
	}

	// move each part to save dir
	for _, downloadRange := range s.Range {
		err := os.Rename(downloadRange.Path, filepath.Join(folder, filepath.Base(downloadRange.Path)))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// save state info to yaml
	marshal, err := yaml.Marshal(s)
	if err != nil {
		return errors.WithStack(err)
	}

	// write state info
	err = ioutil.WriteFile(filepath.Join(folder, consts.StateFile), marshal, 0644)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}

// Read read download state from state yaml file
func Read(task string) (*State, error) {
	file := filepath.Join(os.Getenv("HOME"), consts.SaveFolder, task, consts.StateFile)
	fmt.Printf("Reading state from %s\n", file)

	var err error

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	state := &State{}
	if err = yaml.Unmarshal(bytes, state); err != nil {
		return nil, errors.WithStack(err)
	}

	return state, nil

}

func Resume(task string) (*State, error) {
	return Read(task)
}