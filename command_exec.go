package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/go-zero-boilerplate/loggers"

	"github.com/golang-devops/exec-logger/exec_logger_constants"
)

func NewCommandExecer(logger loggers.LoggerStdIO, logFilePath string, stdErrIsError bool, timeoutKillDuration time.Duration, runArgs []string) *commandExecer {
	aliveFilePath := exec_logger_constants.ALIVE_FILE_NAME
	exitedFilePath := exec_logger_constants.EXITED_FILE_NAME
	mustAbortFilePath := exec_logger_constants.MUST_ABORT_FILE_NAME
	statusHandler := &execStatusHandler{aliveFilePath: aliveFilePath, exitedFilePath: exitedFilePath, mustAbortFilePath: mustAbortFilePath}

	return &commandExecer{
		logger:              logger,
		logFilePath:         logFilePath,
		stdErrIsError:       stdErrIsError,
		timeoutKillDuration: timeoutKillDuration,
		runArgs:             runArgs,
		statusHandler:       statusHandler,
		stdioHandler:        nil, //Set inside `Run` method
	}
}

type commandExecer struct {
	logger              loggers.LoggerStdIO
	logFilePath         string
	stdErrIsError       bool
	timeoutKillDuration time.Duration
	runArgs             []string
	statusHandler       *execStatusHandler
	stdioHandler        *stdioHandler
}

func (c *commandExecer) abortProcess(cmd *exec.Cmd) {
	defer func() {
		if rec := recover(); rec != nil {
			c.stdioHandler.writeErrorLine(fmt.Sprintf("Kill process attempt recovered, recovery: %+v", rec))
		}
	}()

	pid := cmd.Process.Pid
	force := true
	if killErr := KillProcessTree(pid, force); killErr != nil { //if killErr := cmd.Process.Kill(); killErr != nil {
		c.stdioHandler.writeErrorLine(fmt.Sprintf("Cannot kill process with PID %d, error: %s", pid, killErr.Error()))
	}
	c.stdioHandler.writeFileLine(fmt.Sprintf("Successfully killed process with PID %d", pid))
}

func (c *commandExecer) cleanupBeforeStarting() error {
	if err := os.Remove(c.statusHandler.aliveFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Cannot remove alive file '%s', error: %s", c.statusHandler.aliveFilePath, err.Error())
		}
	}
	if err := os.Remove(c.statusHandler.exitedFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Cannot remove exited file '%s', error: %s", c.statusHandler.exitedFilePath, err.Error())
		}
	}
	if err := os.Remove(c.statusHandler.mustAbortFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Cannot remove must-abort file '%s', error: %s", c.statusHandler.mustAbortFilePath, err.Error())
		}
	}
	return nil
}

func (c *commandExecer) runCommand() (exitCode int, returnErr error) {
	if err := c.cleanupBeforeStarting(); err != nil {
		return -1, err
	}

	cmd := exec.Command(c.runArgs[0], c.runArgs[1:]...)

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
	}(c.statusHandler)

	go func(sh *execStatusHandler) {
		for {
			if mustAbort, checkErr := sh.CheckMustAbort(); checkErr != nil {
				c.stdioHandler.writeFileLine(fmt.Sprintf("Unable to check for abort request, error: %s", checkErr.Error()))
			} else if mustAbort {
				c.stdioHandler.writeFileLine("Got ABORT message")
				c.abortProcess(cmd)
				break
			}
			time.Sleep(2 * time.Second)
		}
	}(c.statusHandler)

	c.stdioHandler.stdoutScanner = bufio.NewScanner(stdout)
	c.stdioHandler.stderrScanner = bufio.NewScanner(stderr)

	var wg sync.WaitGroup
	wg.Add(2)
	go c.stdioHandler.startScanningStdout(&wg)
	go c.stdioHandler.startScanningStderr(&wg)

	var waitErr error
	timeoutOccurred := false
	if c.timeoutKillDuration > 0 {
		c.stdioHandler.writeFileLine(fmt.Sprintf("Using timeout of '%s' for process", c.timeoutKillDuration.String()))

		done := make(chan error)
		go func() { done <- cmd.Wait() }()
		select {
		case waitErr = <-done:
			waitErr = waitErr
		case <-time.After(c.timeoutKillDuration):
			c.stdioHandler.writeFileLine(fmt.Sprintf("Timeout of %s reached, now aborting", c.timeoutKillDuration.String()))
			c.abortProcess(cmd)
			timeoutOccurred = true
		}
	} else {
		c.stdioHandler.writeFileLine("No timeout set for process")
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
		return -1, fmt.Errorf("The command timed out after '%s'", c.timeoutKillDuration.String())
	}

	if c.stdioHandler.commandHadStdErr && c.stdErrIsError {
		return -1, fmt.Errorf("The command finished running but had error lines (written to stderr).")
	}

	return 0, nil
}

func (c *commandExecer) Run() (exitCode int, returnErr error) {
	err := os.Remove(c.logFilePath)
	if err != nil && !os.IsNotExist(err) {
		return -1, err
	}

	logFile, err := os.OpenFile(c.logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0655)
	if err != nil {
		return -1, err
	}
	defer logFile.Close()

	c.stdioHandler = &stdioHandler{
		logger: c.logger,
		writer: logFile,
	}

	startTime := time.Now()

	c.stdioHandler.writeFileLine(fmt.Sprintf("Calling commandline: %s", joinCommandLine(c.runArgs)))
	exitCode, err = c.runCommand()

	exitCodeMsg := fmt.Sprintf("Command exited with code %d", exitCode)
	if exitCode != 0 {
		c.stdioHandler.writeErrorLine(exitCodeMsg)
	} else {
		c.stdioHandler.writeFileLine(exitCodeMsg)
	}

	totalDuration := time.Now().Sub(startTime)
	c.statusHandler.WriteExitedJson(exitCode, err, totalDuration)

	c.stdioHandler.writeFileLine(fmt.Sprintf("Total duration was %s", totalDuration.String()))
	if err != nil {
		returnErr = fmt.Errorf("Unable to run command, error: %s", err.Error())
		c.stdioHandler.writeErrorLine(returnErr.Error())
		return exitCode, returnErr
	}

	return 0, nil
}
