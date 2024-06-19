package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/4lxprime/gitdl"
)

func main() {
	repoFlag := flag.String("repo", "", "the repo you want to download folder from")
	folderFlag := flag.String("folder", "/", "folder you want to download from repo")
	outputFlag := flag.String("output", "", "local output folder")
	logsFlag := flag.Bool("logs", true, "display downloading logs")

	flag.Parse()

	if *repoFlag == "" {
		log.Fatal(fmt.Errorf("you should specify the repo you want to clone (e.g. gitdl -folder=/ -output=gitdl 4lxprime/gitdl)"))
	}

	if !strings.Contains(*repoFlag, "/") {
		log.Fatal(fmt.Errorf("repo argument must contains / (e.g. 4lxprime/gitdl or github.com/4lxprime/gitdl)"))
	}

	if *outputFlag == "" {
		repoSplit := strings.Split(*repoFlag, "/")
		repoName := repoSplit[len(repoSplit)-1]

		*outputFlag = repoName
	}

	var options []gitdl.Option
	if *logsFlag {
		options = append(options, gitdl.WithLogs)
	}

	if err := gitdl.DownloadGit(
		*repoFlag,
		*folderFlag,
		*outputFlag,
		options...,
	); err != nil {
		log.Fatal(err)
	}

	if *logsFlag {
		log.Println("download completed successfully")
	}
}
