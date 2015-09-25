package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
)

func init() {
	configPath = "/tmp/.gitwork"
	configFile = "config"
	err := createConfigFile()
	if err != nil {
		logrus.Fatal(err)
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
