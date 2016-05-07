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
	Version = "0.0.0.5"
)

var (
	versionFlag             = flag.Bool("version", false, "Print the version and exit")
	taskFlag                = flag.String("task", "", "The task to run ("+strings.Join(getTaskNamesForFlagHelp(), ", ")+")")
	stdErrIsError           = flag.Bool("stderr-is-error", false, "If any stderr line is printed we will exit with non-zero exit code")
	timeoutKillDuration     = flag.Duration("timeout-kill", 0, "The timeout after which to auto-kill the running process")
	parseErrorPatternsFlag  = flag.String("parse_patterns", "", `Additional error patterns. Split multiple with `+splitParsePatternString+`, for example (without quotes). 'ERROR: (.*)'`+splitParsePatternString+`'MYERROR: (.*)'`)
	recordResourceUsageFlag = flag.Bool("record-resource-usage", false, "Record resource usage - CPU, RAM, etc")
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
	stdioLogger := NewStdioLogger()
	args := flag.Args()
	execer := NewCommandExecer(stdioLogger, *stdErrIsError, *timeoutKillDuration, *recordResourceUsageFlag, args)
	exitCode, err := execer.Run()

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

	err := handleParseLogToStdioCommand(stdioLogger)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		os.Exit(0)
	}

	fmt.Println(fmt.Sprintf("Running version %s", Version))

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
