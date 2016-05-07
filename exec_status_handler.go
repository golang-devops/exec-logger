package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"
	"github.com/golang-devops/exec-logger/resource_usage"
)

type execStatusHandler struct {
	localContextFilePath        string
	aliveFilePath               string
	exitedFilePath              string
	mustAbortFilePath           string
	recordResourceUsageFilePath string
}

func (e *execStatusHandler) writeFile(filePath string, content []byte, mustAppend bool) error {
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("Cannot create parent dir '%s' (of file '%s'), error: %s", parentDir, filePath, err.Error())
	}

	if !mustAppend {
		return ioutil.WriteFile(filePath, content, 0600)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("Unable to open file '%s', error: %s", filePath, err.Error())
	}
	defer file.Close()
	if _, err = file.Write(content); err != nil {
		return fmt.Errorf("Cannot append/write file content, error: %s", err.Error())
	}
	return nil
}

func (e *execStatusHandler) writeJsonFile(filePath string, data interface{}, append bool) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Unable to Marshal data to json. ERROR: %s. Data: %+v.", err.Error(), data)
	}

	return e.writeFile(filePath, jsonBytes, append)
}

func (e *execStatusHandler) WriteLocalContextFile() error {
	data := &exec_logger_dtos.LocalContextDto{}

	if currentUser, err := user.Current(); err != nil {
		data.UserName = fmt.Sprintf("ERROR: Cannot obtain UserName - error '%s'", err.Error())
	} else {
		data.UserName = currentUser.Username
	}

	if hostName, err := os.Hostname(); err != nil {
		data.HostName = fmt.Sprintf("ERROR: Cannot obtain HostName - error '%s'", err.Error())
	} else {
		data.HostName = hostName
	}

	return e.writeJsonFile(e.localContextFilePath, data, false)
}

func (e *execStatusHandler) WriteAlive() error {
	nowTime := time.Now().UTC().Format(exec_logger_constants.ALIVE_TIME_FORMAT)
	return e.writeFile(e.aliveFilePath, []byte(nowTime), false)
}

func (e *execStatusHandler) WriteResourceUsage(procId int) error {
	resourceUsageDTO := &exec_logger_dtos.ResourceUsageDto{}
	fillWarnings := resource_usage.FillResourceUsage(resourceUsageDTO, procId)

	fillWarningsMsgPart := ""
	if len(fillWarnings) > 0 {
		fillWarningsMsgPart = "Warnings while fetching resource usages: " + strings.Join(fillWarnings, "\\n")
	}

	jsonBytes, err := json.Marshal(resourceUsageDTO)
	if err != nil {
		return fmt.Errorf("Cannot marshal resource usage {%+v} to json, error: %s. %s", resourceUsageDTO, err.Error(), fillWarningsMsgPart)
	}
	fileContent := string(jsonBytes) + "\n"
	if err = e.writeFile(e.recordResourceUsageFilePath, []byte(fileContent), true); err != nil {
		return fmt.Errorf("Unable to write resouce-usage file, error: %s. %s", err.Error(), fillWarningsMsgPart)
	}

	//Convert warnings into an error because we successfully wrote the resource file and everything else already.
	if len(fillWarnings) > 0 {
		return fmt.Errorf("Everything succeeded but had warnings. So are now shown as an error. %s", fillWarningsMsgPart)
	}

	return nil
}

func (e *execStatusHandler) WriteExitedJson(exitCode int, err error, duration time.Duration) error {
	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	data := &exec_logger_dtos.ExitStatusDto{
		ExitCode: exitCode,
		Error:    errorStr,
		ExitTime: time.Now().UTC(),
		Duration: duration.String(),
	}

	return e.writeJsonFile(e.exitedFilePath, data, false)
}

func (e *execStatusHandler) CheckMustAbort() (bool, error) {
	_, err := ioutil.ReadFile(e.mustAbortFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
