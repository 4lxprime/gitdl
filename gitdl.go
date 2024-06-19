package gitdl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

type Options struct {
	branch        *string
	authToken     *string
	excludesFiles []string
	replaceMap    Map
	logs          bool
}

type Option func(*Options)

type Map = map[string]string

var defaultBranch string = "main"

var (
	WithBranch = func(branchName string) Option {
		return func(o *Options) {
			o.branch = &branchName
		}
	}
	WithAuth = func(authToken string) Option {
		return func(o *Options) {
			o.authToken = &authToken
		}
	}
	WithExclusions = func(exclusions ...string) Option {
		return func(o *Options) {
			o.excludesFiles = exclusions
		}
	}
	WithReplace = func(replaceMap Map) Option {
		return func(o *Options) {
			o.replaceMap = replaceMap
		}
	}
	WithLogs = func(o *Options) {
		o.logs = true
	}
)

func downloadFile(url, filePath string, opts *Options) error {
	log.Println(url)
	_ = opts

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if len(opts.replaceMap) > 0 {
		// here come the tricky part:

		dataBytes, _ := io.ReadAll(res.Body)

		for replaceKey, replaceValue := range opts.replaceMap {
			// rewriting into buffer remplaced values
			dataBytes = bytes.ReplaceAll(
				dataBytes,
				[]byte(replaceKey),
				[]byte(replaceValue),
			)
		}

		if _, err := file.Write(dataBytes); err != nil {
			return err
		}

		return nil
	}

	_, err = io.Copy(file, res.Body)
	return err
}

func downloadFolder(gitRepo, gitPath, localPath string, opts *Options) error {
	repoApiUrl, err := url.JoinPath(
		"https://api.github.com/repos/",
		gitRepo,
		"contents",
		gitPath,
	)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, repoApiUrl, http.NoBody)
	if err != nil {
		return err
	}

	// adding to the request the github auth token if given
	if opts.authToken != nil {
		req.Header.Add(
			"Authorization",
			fmt.Sprintf(
				"Bearer %s",
				*opts.authToken,
			),
		)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 403 {
		return fmt.Errorf(`github api rate limit exceeded, consider using gitdl.WithAuth("GH_API_KEY") to get a higher rate limit`)

	} else if res.StatusCode != 200 {
		return fmt.Errorf("encountered an error with request: %s", repoApiUrl)
	}

	var items []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return err
	}

	for _, item := range items {
		name := item["name"].(string)
		path := item["path"].(string)
		typeStr := item["type"].(string)

		localItemPath := fmt.Sprintf("%s/%s", localPath, name)

		if opts.logs {
			log.Println("downloading", typeStr, localItemPath)
		}

		// if filepath does match any exclude pattern, we'll just return
		ignore := gitignore.CompileIgnoreLines(opts.excludesFiles...)
		if ok := ignore.MatchesPath(localItemPath); ok {
			return nil
		}

		switch typeStr {
		case "dir":
			if err := os.MkdirAll(localItemPath, os.ModePerm); err != nil {
				return err
			}

			downloadFolder(
				gitRepo,
				path, // next path to download
				localItemPath,
				opts,
			) // reccursive download

		case "file":
			fileURL := fmt.Sprintf(
				"https://raw.githubusercontent.com/%s/%s/%s",
				gitRepo,
				*opts.branch,
				path,
			)

			if err := downloadFile(
				fileURL,
				localItemPath,
				opts,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func DownloadGit(gitRepo, gitPath, localPath string, options ...Option) error {
	opts := new(Options)

	for _, opt := range options {
		opt(opts)
	}

	// setting default branch if no specific branch given
	if opts.branch == nil {
		opts.branch = &defaultBranch
	}

	if strings.Contains(gitRepo, "https://") {
		return fmt.Errorf("repo should not be an url (e.g. 4lxprime/gitdl or github.com/4lxprime/gitdl)")
	}

	// removing the github url
	if strings.Contains(gitRepo, "github.com/") {
		gitRepo = strings.ReplaceAll(gitRepo, "github.com/", "")
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		if err := os.Mkdir(localPath, 0755); err != nil {
			return err
		}
	}

	return downloadFolder(gitRepo, gitPath, localPath, opts)
}
