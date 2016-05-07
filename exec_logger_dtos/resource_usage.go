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

//GetSummedProcessesResourceUsage will just sum the values of all entries and return a single ProcessResourceUsage
func (r *ResourceUsageDto) GetSummedProcessesResourceUsage() *ProcessResourceUsage {
	p := &ProcessResourceUsage{}
	for _, r := range r.ProcessesResourceUsage {
		p.MemoryKB += r.MemoryKB
		p.CPUSeconds += r.CPUSeconds
	}
	return p
}

//ProcessResourceUsage contains resource usage for a single process
type ProcessResourceUsage struct {
	Pid        int
	MemoryKB   int
	CPUSeconds int
}
