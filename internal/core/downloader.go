package core

import (
	"crypto/tls"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/pkg/errors"
	"github.com/sado0823/downloader/internal/consts"
	"github.com/sado0823/downloader/internal/tool"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Downloader struct {
	Url            string
	FileName       string
	Part           int64
	Len            int64
	IPs            []string
	SkipTls        bool
	DownloadRanges []DownloadRange
	Reusable       bool
}

const (
	skipTLS = true
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = &http.Client{Transport: tr}
)

func NewDownloader(rawURL string, part int64) (*Downloader, error) {

	reusable := true

	// check url
	parse, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ips, err := net.LookupIP(parse.Host)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// to ipv4
	ipv4s := tool.GetIPV4s(ips)
	fmt.Printf("downloading from url hosts: %s \n", strings.Join(ipv4s, "|"))

	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.Header.Get(consts.AcceptRange) == "" {
		fmt.Printf("don't set download range, start from 1 \n")
		part = 1
	}

	contentLen := resp.Header.Get(consts.ContentLength)
	if contentLen == "" {
		fmt.Printf("can't get content-length, set parallel = 1")
		reusable = false
		contentLen = "1"
		part = 1
	}

	fmt.Printf("Start downloading with %d connections \n", part)

	intLen, err := strconv.ParseInt(contentLen, 10, 64)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sizeInMb := float64(intLen) / (1024 * 1024)
	if contentLen == "1" {
		fmt.Printf("Download size: not specified\n")
	} else if sizeInMb < 1024 {
		fmt.Printf("Download target size: %.1f MB\n", sizeInMb)
	} else {
		fmt.Printf("Download target size: %.1f GB\n", sizeInMb/1024)
	}

	downloader := &Downloader{
		Url:      rawURL,
		FileName: filepath.Base(rawURL),
		Part:     part,
		Len:      intLen,
		IPs:      ipv4s,
		SkipTls:  skipTLS,
		Reusable: reusable,
	}

	downloadRanges, err := partCalculate(part, intLen, rawURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	downloader.DownloadRanges = downloadRanges

	return downloader, nil
}

func partCalculate(par int64, len int64, url string) ([]DownloadRange, error) {
	var ret []DownloadRange

	for i := int64(0); i < par; i++ {
		from := (len / par) * i
		to := int64(0)
		if i < par-1 {
			to = (len/par)*(i+1) - 1
		} else {
			to = len
		}

		file := filepath.Base(url)
		folder, err := tool.DownloaderFolder(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if err := tool.Mkdir(folder); err != nil {
			return nil, errors.WithStack(err)
		}

		fName := fmt.Sprintf("%s.part%d", file, i)
		path := filepath.Join(folder, fName)
		ret = append(ret, DownloadRange{
			URL:  url,
			Path: path,
			From: from,
			To:   to,
		})
	}
	return ret, nil
}

func (d *Downloader) Downloading(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool, stateSaveChan chan DownloadRange) {
	var bars []*pb.ProgressBar
	var barPool *pb.Pool
	var err error

	if tool.IsTerminal() {
		bars = []*pb.ProgressBar{}
		for i, part := range d.DownloadRanges {
			newBar := pb.New64(part.To - part.From).SetUnits(pb.U_BYTES).Prefix(fmt.Sprintf("%s - %d", d.FileName, i))
			newBar.ShowBar = false
			bars = append(bars, newBar)
		}
		barPool, err = pb.StartPool(bars...)
		if err != nil {
			errorChan <- errors.WithStack(err)
			return
		}
	}

	// Parallel download
	ws := new(sync.WaitGroup)
	for i, p := range d.DownloadRanges {
		ws.Add(1)
		go func(d *Downloader, loop int64, part DownloadRange) {
			defer ws.Done()
			bar := new(pb.ProgressBar)

			if tool.IsTerminal() {
				bar = bars[loop]
			}

			ranges := ""
			if part.To != d.Len {
				ranges = fmt.Sprintf("bytes=%d-%d", part.From, part.To)
			} else {
				ranges = fmt.Sprintf("bytes=%d-", part.From) // last range
			}

			req, err := http.NewRequest(http.MethodGet, d.Url, nil)
			if err != nil {
				errorChan <- errors.WithStack(err)
				return
			}

			if d.Part > 1 {
				req.Header.Add("Range", ranges)
				if err != nil {
					errorChan <- errors.WithStack(err)
					return
				}
			}

			resp, err := client.Do(req)
			if err != nil {
				errorChan <- errors.WithStack(err)
				return
			}
			defer resp.Body.Close()

			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				errorChan <- errors.WithStack(err)
				return
			}

			defer func() {
				err = f.Close()
				if err != nil {
					errorChan <- errors.WithStack(err)
					return
				}
			}()

			var writer io.Writer
			if tool.IsTerminal() {
				writer = io.MultiWriter(f, bar)
			} else {
				writer = io.MultiWriter(f)
			}

			// copy 100 bytes each loop
			current := int64(0)
			for {
				select {
				case <-interruptChan:
					stateSaveChan <- DownloadRange{
						URL:  d.Url,
						Path: part.Path,
						From: current + part.From,
						To:   part.To,
					}
					return
				default:
					written, err := io.CopyN(writer, resp.Body, 100)
					current += written
					if err != nil {
						if err != io.EOF {
							errorChan <- errors.WithStack(err)
							return
						}
						fileChan <- part.Path
						return
					}
				}
			}

		}(d, int64(i), p)
	}

	ws.Wait()

	err = barPool.Stop()
	if err != nil {
		errorChan <- errors.WithStack(err)
		return
	}

	doneChan <- true
}
