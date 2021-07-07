package main

import (
	"github.com/sado0823/downloader/cmd/downloader"
	"log"
)

func main() {
	if err := downloader.RootCmd.Execute(); err != nil {
		log.Fatalf("error occured: %v", err)
	}
}
