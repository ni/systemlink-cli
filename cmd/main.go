package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/ni/systemlink-cli/internal/commandline"
	"github.com/ni/systemlink-cli/internal/model"
	"github.com/ni/systemlink-cli/internal/niservice"
	"github.com/ni/systemlink-cli/internal/parser"
)

const configFileName = "systemlink.yaml"

func getHomeDir() string {
	homeDirPath, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return homeDirPath
}

func loadConfig() commandline.Config {
	homeDirPath := getHomeDir()
	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	directories := []string{homeDirPath, currentDir}

	for _, directory := range directories {
		configFilePath := filepath.Join(directory, configFileName)
		if configFileContent, err := ioutil.ReadFile(configFilePath); err == nil {
			return commandline.NewConfig(configFileContent, directory)
		}
	}
	return commandline.NewConfig([]byte{}, "")
}

func readModels() []model.Data {
	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	modelsDir := filepath.Join(currentDir, "models")

	files, _ := ioutil.ReadDir(modelsDir)
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No model files found. Make sure that the models folder contains json files: %s\n", modelsDir)
		os.Exit(1)
	}
	models := make([]model.Data, len(files))
	for i, f := range files {
		name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		filePath := filepath.Join(modelsDir, f.Name())
		raw, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err)
		}
		models[i] = model.Data{Name: name, Content: raw}
	}
	return models
}

func main() {
	models := readModels()
	config := loadConfig()
	c := commandline.CLI{
		Parser:    parser.SwaggerParser{},
		Service:   niservice.NIService{},
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
		Config:    config,
	}
	c.Exec(os.Args, models)
}
