package integration_test

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
	if !strings.Contains(output, "alarms") {
		t.Errorf("Alarm Service was not registered: %s", output)
	}
	if !strings.Contains(output, "messages") {
		t.Errorf("Message Service was not registered: %s", output)
	}
	if !strings.Contains(output, "taghistory") {
		t.Errorf("Tag Historian Service was not registered: %s", output)
	}
	if !strings.Contains(output, "tagrules") {
		t.Errorf("Tag Rule Engine Service was not registered: %s", output)
	}
	if !strings.Contains(output, "tags") {
		t.Errorf("Tag Service was not registered: %s", output)
	}
	if !strings.Contains(output, "tdms") {
		t.Errorf("TDM Reader Service was not registered: %s", output)
	}
	if !strings.Contains(output, "tests") {
		t.Errorf("Test Monitor Service was not registered: %s", output)
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
