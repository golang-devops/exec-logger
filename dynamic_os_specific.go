package main

import (
	"runtime"
	"strings"
)

var (
	NEWLINE string
)

func init() {
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		NEWLINE = "\r\n"
	default:
		NEWLINE = "\n"
	}
}
