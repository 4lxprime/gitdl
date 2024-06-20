package gitdl

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
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
	noChecksum    bool
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
	WithoutChecksum = func(o *Options) {
		o.noChecksum = true
	}
)

// download file will download
func downloadFile(url, filePath, fileHash string, fileSize int, opts *Options) error {
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
		// fixed size download, like this we save memory
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		if len(bodyBytes) != fileSize {
			return fmt.Errorf("download error: file size don't match, the file may have been corrupted, please retry")
		}

		if !opts.noChecksum { // by default true, so checksum by default
			header := fmt.Sprintf("blob %d\000", len(bodyBytes))
			content := []byte(header + string(bodyBytes))

			hashBytes := sha1.Sum(content)
			hash := hex.EncodeToString(hashBytes[:])

			if hash != fileHash {
				return fmt.Errorf("checksum error: bad checksum (%s)", url)
			}
		}

		for replaceKey, replaceValue := range opts.replaceMap {
			// rewriting into buffer remplaced values
			bodyBytes = bytes.ReplaceAll(
				bodyBytes,
				[]byte(replaceKey),
				[]byte(replaceValue),
			)
		}

		if _, err := file.Write(bodyBytes); err != nil {
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
		itemName := item["name"].(string)
		itemPath := item["path"].(string)
		itemType := item["type"].(string)
		itemHash := item["sha"].(string)
		itemSize := int(item["size"].(float64))

		localItemPath := fmt.Sprintf("%s/%s", localPath, itemName)

		if opts.logs {
			log.Println("downloading", itemType, localItemPath)
		}

		// if filepath does match any exclude pattern, we'll just return
		ignore := gitignore.CompileIgnoreLines(opts.excludesFiles...)
		if ok := ignore.MatchesPath(localItemPath); ok {
			continue
		}

		switch itemType {
		case "dir":
			if err := os.MkdirAll(localItemPath, os.ModePerm); err != nil {
				return err
			}

			if err := downloadFolder( // reccursive download
				gitRepo,
				itemPath, // next path to download
				localItemPath,
				opts,
			); err != nil {
				return err
			}

		case "file":
			fileURL := fmt.Sprintf(
				"https://raw.githubusercontent.com/%s/%s/%s",
				gitRepo,
				*opts.branch,
				itemPath,
			)

			if err := downloadFile(
				fileURL,
				localItemPath,
				itemHash,
				itemSize,
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

	if opts.noChecksum && opts.logs {
		log.Println("checksum disabled: downloaded files can be corrupted")
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		if err := os.Mkdir(localPath, 0755); err != nil {
			return err
		}
	}

	return downloadFolder(gitRepo, gitPath, localPath, opts)
}
