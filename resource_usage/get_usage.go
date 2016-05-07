package resource_usage

import (
	"fmt"
	"time"

	"github.com/go-zero-boilerplate/osvisitors"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"
	"github.com/golang-devops/exec-logger/process_tree"
)

//FillResourceUsage will fill in all the
func FillResourceUsage(dto *exec_logger_dtos.ResourceUsageDto, procId int) (warnings []string) {
	procTree, err := process_tree.LoadProcessTree(procId)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Cannot get resource usage of pid %d, error: %s", procId, err.Error()))
	} else {
		dto.ProcessTree = procTree
	}

	dto.Time = time.Now()

	runtimeOsType, err := osvisitors.GetRuntimeOsType()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Cannot determine runtime os type, error: %s", err.Error()))
	} else {
		helper := NewHelperFromOsType(runtimeOsType)

		cpuPercentage, err := helper.CPUPercentage()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Cannot get CPU percentage"))
		} else {
			dto.CPUPercentage = cpuPercentage
		}

		freePhysicalMemKB, err := helper.FreePhysicalMemoryKB()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Cannot get CPU free physical memory"))
		} else {
			dto.FreePhysicalMemoryKB = freePhysicalMemKB
		}

		freeVirtualMemKB, err := helper.FreeVirtualMemoryKB()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Cannot get CPU free virtual memory"))
		} else {
			dto.FreeVirtualMemoryKB = freeVirtualMemKB
		}

		//It might have been skipped due to error
		if dto.ProcessTree != nil {
			processResources := []*exec_logger_dtos.ProcessResourceUsage{}

			allPids := dto.ProcessTree.FlattenedPids()
			for _, pid := range allPids {
				memKB, cpuDuration, err := helper.ProcessUsedCPUAndMemoryKB(pid)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("Cannot get CPU+Mem for pid %d, error: %s", pid, err.Error()))
					continue
				}
				processResources = append(processResources, &exec_logger_dtos.ProcessResourceUsage{
					Pid:        pid,
					MemoryKB:   memKB,
					CPUSeconds: int(cpuDuration.Seconds()),
				})
			}

			dto.ProcessesResourceUsage = processResources
		}
		// helper.ProcessUsedCPUAndMemoryKB(pid int)
	}

	return
}
