package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/libgit2/git2go"
)

var (
	gitPathEnv   = os.Getenv("GIT")
	gitReposPath = flag.String("g", gitPathEnv, "Path to git repositories")
)

type gitinfo struct {
	Name string
	Date time.Time
}

// get the repositories list
func getRepositories(gitpath string) ([]string, error) {
	repos := make([]string, 0)
	if _, err := os.Stat(gitpath); os.IsNotExist(err) {
		return nil, err
	}

	// path is absolute
	absPath, err := filepath.Abs(gitpath)
	if err != nil {
		return nil, err
	}

	// read the files on the git path
	files, err := ioutil.ReadDir(gitpath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		absGitPath := path.Join(absPath, f.Name())
		repos = append(repos, absGitPath)
	}
	return repos, nil
}

func gitRepository(repo string) (*gitinfo, error) {
	r, err := git.OpenRepository(repo)
	if err != nil {
		return nil, err
	}

	head, err := r.Head()
	if err != nil {
		return nil, err
	}
	// Last commit form HEAD, do not assume
	// HEAD points to master.
	lastCommit, err := r.LookupCommit(head.Target())
	if err != nil {
		return nil, err
	}
	author := lastCommit.Author()

	gitInfo := &gitinfo{
		Date: author.When,
	}
	return gitInfo, nil
}

func main() {
	flag.Parse()
	gitInfoRepo := make([]*gitinfo, 0)
	if *gitReposPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	repos, err := getRepositories(*gitReposPath)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, r := range repos {
		info, err := gitRepository(r)
		if err != nil {
			continue
		}
		indexLastSlash := strings.LastIndex(r, "/") + 1
		info.Name = r[indexLastSlash:]
		gitInfoRepo = append(gitInfoRepo, info)
	}
}
