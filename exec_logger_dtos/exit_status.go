package exec_logger_dtos

import (
	"strings"

	"time"
)

type ExitStatusDto struct {
	ExitCode int
	Error    string
	ExitTime time.Time
	Duration string
}

func (e *ExitStatusDto) HasError() bool {
	return strings.TrimSpace(e.Error) != ""
}
