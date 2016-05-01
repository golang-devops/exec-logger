package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"
)

type execStatusHandler struct {
	aliveFilePath     string
	exitedFilePath    string
	mustAbortFilePath string
}

func (e *execStatusHandler) WriteAlive() error {
	nowTime := time.Now().UTC().Format(exec_logger_constants.ALIVE_TIME_FORMAT)
	return ioutil.WriteFile(e.aliveFilePath, []byte(nowTime), 0655)
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

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Unable to Marshal data to json. ERROR: %s. Data: %+v.", err.Error(), data)
	}

	return ioutil.WriteFile(e.exitedFilePath, jsonBytes, 0655)
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
