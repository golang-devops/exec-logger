package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	taskFlag               = flag.String("task", "", "The task to run ("+strings.Join(getTaskNamesForFlagHelp(), ", ")+")")
	stdErrIsError          = flag.Bool("stderr-is-error", false, "If any stderr line is printed we will exit with non-zero exit code")
	logFileFlag            = flag.String("logfile", "", "The path to the log file")
	timeoutKillDuration    = flag.Duration("timeout-kill", 0, "The timeout after which to auto-kill the running process")
	parseErrorPatternsFlag = flag.String("parse_patterns", "", `Additional error patterns. Split multiple with `+splitParsePatternString+`, for example (without quotes). 'ERROR: (.*)'`+splitParsePatternString+`'MYERROR: (.*)'`)
)

var (
	splitParsePatternString = "[{|}]"

	tasks = []struct {
		Name    string
		Handler func()
	}{
		{Name: "exec", Handler: doExecCommand},
		{Name: "parselog", Handler: doParseLogToStdioCommand},
	}
)

func getTaskNamesForFlagHelp() (names []string) {
	for _, t := range tasks {
		names = append(names, t.Name)
	}
	return
}

func doExecCommand() {
	if len(*logFileFlag) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	stdioLogger := NewStdioLogger()
	logFilePath := *logFileFlag
	args := flag.Args()
	exitCode, err := handleExecCommand(stdioLogger, logFilePath, *stdErrIsError, *timeoutKillDuration, args)

	fmt.Printf("exit code was %d\n", exitCode)

	if err != nil {
		log.Printf("Error running command: %s.\nThe arguments used were: %s", err.Error(), fmt.Sprintf("%+v", args))
		if exitCode != -1 {
			os.Exit(exitCode)
		} else {
			os.Exit(2)
		}
	}

	os.Exit(0)
}

func doParseLogToStdioCommand() {
	if len(*logFileFlag) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	stdioLogger := NewStdioLogger()

	additionalErrorPatterns = []*regexp.Regexp{}
	if len(*parseErrorPatternsFlag) > 0 {
		for _, s := range strings.Split(*parseErrorPatternsFlag, splitParsePatternString) {
			if strings.TrimSpace(s) == "" {
				continue
			}
			additionalErrorPatterns = append(additionalErrorPatterns, regexp.MustCompile(s))
			stdioLogger.Out("Additional error pattern added: %s", s)
		}
	}

	logFilePath := *logFileFlag
	err := handleParseLogToStdioCommand(stdioLogger, logFilePath)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	if len(*taskFlag) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	var handler func() = nil
	for _, t := range tasks {
		if strings.EqualFold(t.Name, *taskFlag) {
			handler = t.Handler
		}
	}
	if handler == nil {
		log.Fatalf("Unsupported task '%s'", *taskFlag)
	}
	handler()
}
