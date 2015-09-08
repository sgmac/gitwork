package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var (
	gitPathEnv   = os.Getenv("GIT")
	gitReposPath = flag.String("g", gitPathEnv, "Path to git repositories")
)

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

func main() {
	flag.Parse()
	if *gitReposPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

}
