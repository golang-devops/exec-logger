package exec_logger_dtos

import (
	"time"

	"github.com/golang-devops/exec-logger/process_tree"
)

//ResourceUsageDto holds the resource usage details
type ResourceUsageDto struct {
	Time                   time.Time
	CPUPercentage          int
	FreePhysicalMemoryKB   int
	FreeVirtualMemoryKB    int
	ProcessesResourceUsage []*ProcessResourceUsage
	ProcessTree            *process_tree.ProcessTree
}

//ProcessResourceUsage contains resource usage for a single process
type ProcessResourceUsage struct {
	Pid        int
	MemoryKB   int
	CPUSeconds int
}
