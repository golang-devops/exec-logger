package resource_usage

import (
	"time"

	"github.com/go-zero-boilerplate/osvisitors"
)

//Helper is basically a "facade" to hide the ugly logic
type Helper interface {
	CPUPercentage() (int, error)
	FreePhysicalMemoryKB() (int, error)
	FreeVirtualMemoryKB() (int, error)

	ProcessUsedCPUAndMemoryKB(pid int) (memKB int, cpuDuration time.Duration, returnErr error)
}

//NewHelperFromOsType will create a new Helper from the OsType
func NewHelperFromOsType(osType osvisitors.OsType) Helper {
	visitor := &visitorCreateHelper{}
	osType.Accept(visitor)
	return visitor.helper
}

type visitorCreateHelper struct{ helper Helper }

func (v *visitorCreateHelper) VisitWindows() {
	v.helper = &winHelper{}
}
func (v *visitorCreateHelper) VisitLinux() {
	//TODO: Implement linux, should be able to use https://github.com/shirou/gopsutil for many of the methods
	panic("Linux not yet implemented")
}
func (v *visitorCreateHelper) VisitDarwin() {
	//TODO: Implement darwin, should be able to use https://github.com/shirou/gopsutil for many of the methods
	panic("Darwin not yet implemented")
}
