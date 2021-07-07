package main

import (
	"flag"
	"github.com/sado0823/downloader/internal"
	"github.com/sado0823/downloader/internal/core"
	"github.com/sado0823/downloader/internal/tool"
	"log"
	"runtime"
)

var (
	url   = flag.String("f", "https://download.jetbrains.com/go/goland-2020.2.2.dmg", "the download url")
	count = flag.Int("c", runtime.NumCPU(), "parallel count, default is cpu num")
)

func main() {
	flag.Parse()

	//if *url == "" {
	//	//log.Fatalf("please enter your url")
	//}

	//folder, err := tool.DownloaderFolder(*url)
	//if err != nil {
	//	log.Fatalf("error occured: %v", err)
	//}
	//if tool.FolderExist(folder) {
	//	fmt.Printf("Task already exist, remove it first \n")
	//	folder, err = tool.DownloaderFolder(*url)
	//	if err != nil {
	//		log.Fatalf("error occured: %v", err)
	//	}
	//	if err := os.RemoveAll(folder); err != nil {
	//		log.Fatalf("error occured: %v", err)
	//	}
	//}

	// resume
	task := ""
	if !tool.InvalidURL(*url) {
		task = tool.DownloaderName(*url)
	} else {
		task = *url
	}

	state, err := core.Resume(task)
	if err != nil {
		log.Fatalf("error occured: %v", err)
	}

	err = internal.Do(*url, state, int64(*count))
	if err != nil {
		log.Fatalf("error occured: %v", err)
	}

}
