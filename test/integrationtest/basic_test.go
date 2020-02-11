package main_test

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestAllServicesRegistered(t *testing.T) {
	stdout, err := exec.Command(executableName()).Output()
	if err != nil {
		t.Errorf("Error executing systemlink binary: %s", err.Error())
	}

	output := string(stdout)
	if !strings.Contains(output, "messages") {
		t.Errorf("Message service was not registered: %s", output)
	}
	if !strings.Contains(output, "tags") {
		t.Errorf("Tag service was not registered: %s", output)
	}
}

func TestShowsServiceOperations(t *testing.T) {
	stdout, err := exec.Command(executableName(), "tags").Output()
	if err != nil {
		t.Errorf("Error executing systemlink binary: %s", err.Error())
	}

	output := string(stdout)
	if !strings.Contains(output, "create-tag") {
		t.Errorf("Service operations are not shown: %s", output)
	}
}

func executableName() string {
	if runtime.GOOS == "windows" {
		return "../../build/systemlink.exe"
	}
	if runtime.GOOS == "darwin" {
		return "../../build/systemlink.osx"
	}
	return "../../build/systemlink"
}
