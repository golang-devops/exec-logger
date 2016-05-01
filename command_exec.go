package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-zero-boilerplate/loggers"

	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"
)

type stdioHandler struct {
	sync.RWMutex

	runnerLogger  loggers.LoggerStdIO
	writer        io.Writer
	stdoutScanner *bufio.Scanner
	stderrScanner *bufio.Scanner

	commandHadStdErr bool
}

func (s *stdioHandler) writeFileLine(line string) {
	s.Lock()
	defer s.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err := io.WriteString(s.writer, fmt.Sprintf("[%s] %s%s", timestamp, line, NEWLINE))
	if err != nil {
		s.runnerLogger.Err("Cannot write, error: %s", err.Error())
	}
}

func (s *stdioHandler) writeErrorLine(e string) {
	if strings.TrimSpace(e) == "" {
		return
	}

	s.commandHadStdErr = true
	s.writeFileLine("EASY_EXEC_ERROR: " + e)
}

func (s *stdioHandler) startScanningStdout(wg *sync.WaitGroup) {
	defer wg.Done()
	for s.stdoutScanner.Scan() {
		s.writeFileLine(s.stdoutScanner.Text())
	}
}

func (s *stdioHandler) startScanningStderr(wg *sync.WaitGroup) {
	defer wg.Done()
	for s.stderrScanner.Scan() {
		s.writeErrorLine(s.stderrScanner.Text())
	}
}

type execStatusHandler struct {
	aliveFilePath     string
	exitedFilePath    string
	mustAbortFilePath string
}

func (e *execStatusHandler) WriteAlive() error {
	nowTime := time.Now().UTC().Format(exec_logger_constants.ALIVE_TIME_FORMAT)
	return ioutil.WriteFile(e.aliveFilePath, []byte(nowTime), 0655)
}

func (e *execStatusHandler) WriteExitedJson(exitCode int, err error) error {
	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	data := &exec_logger_dtos.ExitStatusDto{
		ExitCode: exitCode,
		Time:     time.Now().UTC(),
		Error:    errorStr,
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

func abortProcess(stdioHandler *stdioHandler, cmd *exec.Cmd) {
	defer func() {
		if rec := recover(); rec != nil {
			stdioHandler.writeErrorLine(fmt.Sprintf("Kill process attempt recovered, recovery: %+v", rec))
		}
	}()

	pid := cmd.Process.Pid
	force := true
	if killErr := KillProcessTree(pid, force); killErr != nil { //if killErr := cmd.Process.Kill(); killErr != nil {
		stdioHandler.writeErrorLine(fmt.Sprintf("Cannot kill process with PID %d, error: %s", pid, killErr.Error()))
	}
	stdioHandler.writeFileLine(fmt.Sprintf("Successfully killed process with PID %d", pid))
}

func cleanupBeforeStarting(statusHandler *execStatusHandler) error {
	if err := os.Remove(statusHandler.aliveFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Cannot remove alive file '%s', error: %s", statusHandler.aliveFilePath, err.Error())
		}
	}
	if err := os.Remove(statusHandler.exitedFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Cannot remove exited file '%s', error: %s", statusHandler.exitedFilePath, err.Error())
		}
	}
	if err := os.Remove(statusHandler.mustAbortFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Cannot remove must-abort file '%s', error: %s", statusHandler.mustAbortFilePath, err.Error())
		}
	}
	return nil
}

func runCommand(stdioHandler *stdioHandler, statusHandler *execStatusHandler, runArgs []string, stdErrIsError bool, timeoutKillDuration time.Duration) (exitCode int, returnErr error) {
	if err := cleanupBeforeStarting(statusHandler); err != nil {
		return -1, err
	}

	cmd := exec.Command(runArgs[0], runArgs[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return -1, err
	}

	err = cmd.Start()
	if err != nil {
		return -1, err
	}

	go func(sh *execStatusHandler) {
		for {
			sh.WriteAlive()
			time.Sleep(2 * time.Second)
		}
	}(statusHandler)

	go func(sh *execStatusHandler) {
		for {
			if mustAbort, checkErr := sh.CheckMustAbort(); checkErr != nil {
				stdioHandler.writeFileLine(fmt.Sprintf("Unable to check for abort request, error: %s", checkErr.Error()))
			} else if mustAbort {
				stdioHandler.writeFileLine("Got ABORT message")
				abortProcess(stdioHandler, cmd)
				break
			}
			time.Sleep(2 * time.Second)
		}
	}(statusHandler)

	stdioHandler.stdoutScanner = bufio.NewScanner(stdout)
	stdioHandler.stderrScanner = bufio.NewScanner(stderr)

	var wg sync.WaitGroup
	wg.Add(2)
	go stdioHandler.startScanningStdout(&wg)
	go stdioHandler.startScanningStderr(&wg)

	var waitErr error
	timeoutOccurred := false
	if timeoutKillDuration > 0 {
		stdioHandler.writeFileLine(fmt.Sprintf("Using timeout of '%s' for process", timeoutKillDuration.String()))

		done := make(chan error)
		go func() { done <- cmd.Wait() }()
		select {
		case waitErr = <-done:
			waitErr = waitErr
		case <-time.After(timeoutKillDuration):
			stdioHandler.writeFileLine(fmt.Sprintf("Timeout of %s reached, now aborting", timeoutKillDuration.String()))
			abortProcess(stdioHandler, cmd)
			timeoutOccurred = true
		}
	} else {
		stdioHandler.writeFileLine("No timeout set for process")
		waitErr = cmd.Wait()
	}

	//TODO: Just give things time to cool down, like writing of the "Successfully killed process" log. This can however be improved with a WaitGroup
	time.Sleep(500 * time.Millisecond)

	if waitErr != nil {
		if exitCode, ok := getExitCodeFromError(waitErr); ok {
			return exitCode, waitErr
		}
		return -1, waitErr
	}
	wg.Wait()

	if timeoutOccurred {
		return -1, fmt.Errorf("The command timed out after '%s'", timeoutKillDuration.String())
	}

	if stdioHandler.commandHadStdErr && stdErrIsError {
		return -1, fmt.Errorf("The command finished running but had error lines (written to stderr).")
	}

	return 0, nil
}

func joinCommandLine(runArgs []string) string {
	formatted := []string{}
	for _, ra := range runArgs {
		trimmed := strings.Trim(ra, " '\"")
		if strings.Contains(trimmed, " ") {
			formatted = append(formatted, "\""+trimmed+"\"")
		} else {
			formatted = append(formatted, trimmed)
		}
	}
	return strings.Join(formatted, " ")
}

func handleExecCommand(runnerLogger loggers.LoggerStdIO, logFilePath string, stdErrIsError bool, timeoutKillDuration time.Duration, runArgs []string) (exitCode int, returnErr error) {
	err := os.Remove(logFilePath)
	if err != nil && !os.IsNotExist(err) {
		return -1, err
	}

	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0655)
	if err != nil {
		return -1, err
	}
	defer logFile.Close()

	h := &stdioHandler{
		runnerLogger: runnerLogger,
		writer:       logFile,
	}

	aliveFilePath := exec_logger_constants.ALIVE_FILE_NAME
	exitedFilePath := exec_logger_constants.EXITED_FILE_NAME
	mustAbortFilePath := exec_logger_constants.MUST_ABORT_FILE_NAME
	statusHandler := &execStatusHandler{aliveFilePath: aliveFilePath, exitedFilePath: exitedFilePath, mustAbortFilePath: mustAbortFilePath}

	h.writeFileLine(fmt.Sprintf("Calling commandline: %s", joinCommandLine(runArgs)))
	exitCode, err = runCommand(h, statusHandler, runArgs, stdErrIsError, timeoutKillDuration)

	exitCodeMsg := fmt.Sprintf("Command exited with code %d", exitCode)
	if exitCode != 0 {
		h.writeErrorLine(exitCodeMsg)
	} else {
		h.writeFileLine(exitCodeMsg)
	}

	statusHandler.WriteExitedJson(exitCode, err)

	if err != nil {
		returnErr = fmt.Errorf("Unable to run command, error: %s", err.Error())
		h.writeErrorLine(returnErr.Error())
		return exitCode, returnErr
	}

	return 0, nil
}
