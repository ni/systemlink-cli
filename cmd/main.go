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

func loadConfig() (commandline.Config, error) {
	config := commandline.Config{}
	homeDirPath, err := homedir.Dir()
	if err != nil {
		return config, err
	}
	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	directories := []string{homeDirPath, currentDir}

	for _, directory := range directories {
		configFilePath := filepath.Join(directory, configFileName)
		configFileContent, err := ioutil.ReadFile(configFilePath)
		if err == nil {
			return commandline.NewConfig(configFileContent, directory)
		}
	}
	return config, nil
}

func readModels() ([]model.Data, error) {
	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	modelsDir := filepath.Join(currentDir, "models")

	files, _ := ioutil.ReadDir(modelsDir)
	if len(files) == 0 {
		return nil, fmt.Errorf("No model files found. Make sure that the models folder contains swagger yaml files: %s", modelsDir)
	}
	models := make([]model.Data, len(files))
	for i, f := range files {
		name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		filePath := filepath.Join(modelsDir, f.Name())
		raw, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("Error reading file '%s': %v", filePath, err)
		}
		models[i] = model.Data{Name: name, Content: raw}
	}
	return models, nil
}

func main() {
	models, err := readModels()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading models:", err)
		os.Exit(1)
	}
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config:", err)
		os.Exit(1)
	}

	c := commandline.CLI{
		Parser:    parser.SwaggerParser{},
		Service:   niservice.NIService{},
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
		Config:    config,
	}
	_, exitStatus := c.Exec(os.Args, models)
	os.Exit(exitStatus)
}
