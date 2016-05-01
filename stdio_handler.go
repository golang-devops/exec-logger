package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-zero-boilerplate/loggers"
)

type stdioHandler struct {
	sync.RWMutex

	logger        loggers.LoggerStdIO
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
		s.logger.Err("Cannot write, error: %s", err.Error())
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
