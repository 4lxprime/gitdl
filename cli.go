package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

func main() {
	repoFlag := flag.String("repo", "", "[required] the repo you want to download folder from")
	folderFlag := flag.String("folder", "/", "[optional] folder you want to download from repo")
	branchFlag := flag.String("branch", "main", "[optional] the branch you want to clone")
	authFlag := flag.String("auth", "", "[optional] your github auth token")
	outputFlag := flag.String("output", "", "[optional] local output folder")
	logsFlag := flag.Bool("logs", true, "[optional] display downloading logs")
	noChecksumFlag := flag.Bool("nochecksum", false, "[optional] disable downloading checksum")

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

	var options []Option

	if *logsFlag {
		options = append(options, WithLogs)
	}

	if *noChecksumFlag {
		options = append(options, WithoutChecksum)
	}

	if *authFlag != "" {
		options = append(options, WithAuth(*authFlag))
	}

	options = append(options, WithBranch(*branchFlag)) // default set to main

	if err := DownloadGit(
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
