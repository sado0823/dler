package tool

import (
	"fmt"
	"github.com/cheggaaa/pb"
	_ "github.com/cheggaaa/pb"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal/consts"
	"io"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func MergeFile(files []string, to string) error {
	// merge order
	sort.Strings(files)

	bar := new(pb.ProgressBar)
	bar.ShowBar = false

	if IsTerminal() {
		fmt.Printf("mergeing file... \n")
		bar = pb.StartNew(len(files))
	}

	destFile, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return errors.WithStack(err)
	}

	// copy files
	for _, f := range files {
		if err := copyFile(f, destFile); err != nil {
			return err
		}

		if IsTerminal() {
			bar.Increment()
		}

	}

	if IsTerminal() {
		bar.Finish()
	}

	return destFile.Close()
}

func copyFile(file string, to io.Writer) error {
	openFile, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = io.Copy(to, openFile)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}

func InvalidURL(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err != nil
}

func DownloaderName(url string) string {
	return filepath.Base(url)
}

func DownloaderHome() string {
	return path.Join(os.Getenv("HOME"), consts.SaveFolder)

}

func DownloaderFolder(url string) (string, error) {
	fullPath := DownloaderHome()

	absolutePath, err := filepath.Abs(path.Join(fullPath, filepath.Base(url)))
	if err != nil {
		return "", errors.WithStack(err)
	}

	// To prevent path traversal attack
	relative, err := filepath.Rel(fullPath, absolutePath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if strings.Contains(relative, "..") {
		return "", errors.WithStack(errors.New("Your download file may have a path traversal attack"))
	}

	return absolutePath, nil
}

func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

func FolderExist(folder string) bool {
	_, err := os.Stat(folder)
	return err == nil || os.IsExist(err)
}

func Mkdir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		//oldMask := syscall.Umask(0)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		//syscall.Umask(oldMask)
	}
	return nil
}

func GetIPV4(ip net.IP) string {
	i := ip.To4()
	if i != nil {
		return i.String()
	}
	return ""
}

func GetIPV4s(ips []net.IP) []string {
	res := make([]string, 0)

	for _, ip := range ips {
		if v := ip.To4(); v != nil {
			res = append(res, v.String())
		}
	}

	return res
}
