package exec_logger_dtos

import (
	"strings"

	"time"
)

type ExitStatusDto struct {
	ExitCode int
	Time     time.Time
	Error    string
}

func (e *ExitStatusDto) HasError() bool {
	return strings.TrimSpace(e.Error) != ""
}
