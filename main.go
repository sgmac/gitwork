package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/libgit2/git2go"
)

const (
	grey  = "\x1b[1;30m"
	red   = "\x1b[1;31m"
	green = "\x1b[1;32m"
	reset = "\x1b[0m"
)

var (
	gitPathEnv   = os.Getenv("GIT")
	gitReposPath = flag.String("g", gitPathEnv, "Path to git repositories")
	configPath   = path.Join(os.Getenv("HOME"), ".gitwork")
	configFile   = "config"
	activeWork   = 90
)

type gitinfo struct {
	Name   string
	Date   time.Time
	Branch string
}

type gitInfoSorter struct {
	repos []gitinfo
}

type Config struct {
	Global Global
}

type Global struct {
	//TODO: Provide branch option
	DaysAgo int `toml:"daysago"`
}

func (r gitInfoSorter) Len() int {
	return len(r.repos)
}

func (r gitInfoSorter) Swap(i, j int) {
	r.repos[i], r.repos[j] = r.repos[j], r.repos[i]
}

func (r gitInfoSorter) Less(i, j int) bool {
	// Sort from recent to older change
	return r.repos[j].Date.Before(r.repos[i].Date)
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

	branch, err := head.Branch().Name()
	if err != nil {
		return nil, err
	}

	gitInfo := &gitinfo{
		Date:   author.When,
		Branch: branch,
	}
	return gitInfo, nil
}

func listGitInfo(repos gitInfoSorter) {
	var message string
	w := tabwriter.NewWriter(os.Stdout, 25, 8, 0, ' ', 0)
	fmt.Fprintln(w, "GIT\tBRANCH\tLAST CHANGE")

	for _, r := range repos.repos {
		date := time.Now().Sub(r.Date)
		daysAgo := int(date.Hours()) / 24
		message = fmt.Sprintf("%d %sdays ago%s %s(abandoned)%s", daysAgo, grey, reset, red, reset)
		if daysAgo < activeWork {
			message = fmt.Sprintf("%d %sdays ago%s", daysAgo, green, reset)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.Branch, message)
	}
	w.Flush()
}

func createConfigFile() error {
	var config Config
	file := path.Join(configPath, configFile)
	var f *os.File
	// create config path
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := os.Mkdir(configPath, 0755)
		if err != nil {
			return err
		}
	}
	// create config file
	if _, err := os.Stat(file); os.IsNotExist(err) {
		f, err = os.Create(file)
		if err != nil {
			return err
		}
	}
	toml.NewEncoder(f).Encode(&config)
	return nil
}

func main() {
	flag.Parse()
	// check if the config file exits or create an empty one
	err := createConfigFile()
	if err != nil {
		logrus.Fatal(err)
	}

	gitReposInfo := make([]gitinfo, 0)
	if *gitReposPath == "" {
		fmt.Fprintf(os.Stderr, "usage: gitwork [OPTIONS]\n")
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
		gitReposInfo = append(gitReposInfo, *info)
	}
	reposSorted := gitInfoSorter{repos: gitReposInfo}
	sort.Sort(reposSorted)
	listGitInfo(reposSorted)
}
