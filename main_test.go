package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/libgit2/git2go"
)

func init() {
	gitPathEnv = "/tmp/git"
	configPath = "/tmp/.gitwork"
	configFile = "config"
	err := createConfigFile()
	if err != nil {
		logrus.Fatal(err)
	}
}

func createGitRepositories() {

	signature := &git.Signature{
		Name:  "John Doe",
		Email: "johndoe@nowhere.com",
		When:  time.Now(),
	}

	repositories := []struct {
		Repo          string
		CommitMessage string
	}{
		{path.Join(gitPathEnv, "test1"), "First commit"},
		{path.Join(gitPathEnv, "test2"), "Initial commit"},
		{path.Join(gitPathEnv, "nongit"), ""},
	}
	for _, r := range repositories {
		if r.CommitMessage == "" {
			os.Mkdir(r.Repo, 0755)
			continue
		}

		// git init
		gitRepo, err := git.InitRepository(r.Repo, false)
		if err != nil {
			logrus.Fatal("0: ", err)
		}

		// create content for the file
		filename := path.Join(r.Repo, "file")
		data := []byte(`This is a file`)
		err = ioutil.WriteFile(filename, data, 0755)
		if err != nil {
			logrus.Fatal("1: ", err)
		}

		// get repo's index
		index, err := gitRepo.Index()
		if err != nil {
			logrus.Fatal("2: ", err)
		}

		// add content to the index/staging area, file
		// must be relative to the repository.
		err = index.AddByPath("file")
		if err != nil {
			logrus.Fatal("3: ", err)
		}

		treeId, err := index.WriteTree()
		if err != nil {
			logrus.Fatal("4: ", err)
		}

		err = index.Write()
		if err != nil {
			logrus.Fatal("5: ", err)
		}

		tree, err := gitRepo.LookupTree(treeId)
		if err != nil {
			logrus.Fatal("6: ", err)
		}

		// Create commit, as this is the first commit, there are no parents.
		_, err = gitRepo.CreateCommit("refs/heads/master", signature, signature, r.CommitMessage, tree)
		if err != nil {
			logrus.Fatal("7: ", err)
		}
	}
}

func TestReadConfigFile(t *testing.T) {
	// Default flag: activeWork defines how many days
	// if not value is provided should be expected 90.
	configExpected := &Config{Global: Global{DaysAgo: 90}}
	config, err := readConfigFile()
	if err != nil {
		logrus.Fatal(err)
	}

	if configExpected.Global != config.Global {
		fmt.Println("Expected: ", configExpected)
		fmt.Println("Got: ", config)
		t.Error("Default configuration does not match ")
	}

	defer func() {
		os.RemoveAll(configPath)
	}()
}

func TestGetRepositories(t *testing.T) {
	// Create two git repositories and one non git repo. This
	// gets all the directories
	createGitRepositories()
	reposExpected := []string{"/tmp/git/test1", "/tmp/git/test2", "/tmp/git/nongit"}
	repos, err := getRepositories(gitPathEnv)
	if err != nil {
		logrus.Fatal(err)
	}

	if len(reposExpected) != len(repos) {
		t.Error("Repositories not found")
	}
}
