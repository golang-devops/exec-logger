package exec_logger_constants

import "path/filepath"

const (
	__EXEC_LOGGER_FILES_SUBDIR = "exec-logger"
)

var (
	LOG_FILE_NAME           = filepath.Join(__EXEC_LOGGER_FILES_SUBDIR, "log.log")
	LOCAL_CONTEXT_FILE_NAME = filepath.Join(__EXEC_LOGGER_FILES_SUBDIR, "local-context.json")
	ALIVE_FILE_NAME         = filepath.Join(__EXEC_LOGGER_FILES_SUBDIR, "alive.txt")
	EXITED_FILE_NAME        = filepath.Join(__EXEC_LOGGER_FILES_SUBDIR, "exited.json")
	MUST_ABORT_FILE_NAME    = filepath.Join(__EXEC_LOGGER_FILES_SUBDIR, "must-abort.txt")
)
