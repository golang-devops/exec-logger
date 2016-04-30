package main

import (
	"fmt"
	"os"

	"github.com/go-zero-boilerplate/loggers"
)

func NewStdioLogger() loggers.LoggerStdIO {
	return &stdio{}
}

type stdio struct{}

func (s *stdio) Err(format string, args ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func (s *stdio) Out(format string, args ...interface{}) {
	os.Stdout.WriteString(fmt.Sprintf(format, args...) + "\n")
}
