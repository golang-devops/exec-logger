package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"
)

type execStatusHandler struct {
	localContextFilePath string
	aliveFilePath        string
	exitedFilePath       string
	mustAbortFilePath    string
}

func (e *execStatusHandler) writeFile(filePath string, content []byte) error {
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("Cannot create parent dir '%s' (of file '%s'), error: %s", parentDir, filePath, err.Error())
	}
	return ioutil.WriteFile(filePath, content, 0655)
}

func (e *execStatusHandler) writeJsonFile(filePath string, data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Unable to Marshal data to json. ERROR: %s. Data: %+v.", err.Error(), data)
	}

	return e.writeFile(filePath, jsonBytes)
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

	return e.writeJsonFile(e.localContextFilePath, data)
}

func (e *execStatusHandler) WriteAlive() error {
	nowTime := time.Now().UTC().Format(exec_logger_constants.ALIVE_TIME_FORMAT)
	return e.writeFile(e.aliveFilePath, []byte(nowTime))
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

	return e.writeJsonFile(e.exitedFilePath, data)
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
