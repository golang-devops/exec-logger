package main

import (
	"bufio"
	"os"
	"regexp"

	"github.com/go-zero-boilerplate/loggers"
)

var (
	errorLogLinePattern     = regexp.MustCompile(`\[[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}\] EASY_EXEC_ERROR: (.*)`)
	additionalErrorPatterns []*regexp.Regexp
)

func lineHasError(line string) bool {
	if errorLogLinePattern.MatchString(line) {
		return true
	}

	for _, ap := range additionalErrorPatterns {
		if ap.MatchString(line) {
			return true
		}
	}

	return false
}

func handleParseLogToStdioCommand(stdioLogger loggers.LoggerStdIO, logFilePath string) error {
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		txt := scanner.Text()
		if lineHasError(txt) {
			stdioLogger.Err("%s", txt)
		} else {
			stdioLogger.Out("%s", txt)
		}
	}

	return nil
}
