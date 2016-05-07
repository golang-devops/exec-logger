package process_tree

import (
	"fmt"
	"os"

	"github.com/go-zero-boilerplate/wmic-proccess-tree/process"
	process2 "github.com/shirou/gopsutil/process"
)

//Process is just a wrapper around `os.Process` to include children processes
type Process struct {
	*os.Process
	Name       string
	Exe        string
	NumThreads int
	Cmdline    string
	Children   []*Process
}

func (p *Process) loadTreeDetails() error {
	psutilProc, err := process2.NewProcess(int32(p.Pid))
	if err != nil {
		return fmt.Errorf("Cannot load details of pid %d, error: %s", p.Pid, err.Error())
	}

	name, err := psutilProc.Name()
	if err != nil {
		return fmt.Errorf("Cannot load process name of pid %d, error: %s", p.Pid, err.Error())
	}
	p.Name = name

	exe, err := psutilProc.Exe()
	if err != nil {
		return fmt.Errorf("Cannot load process exe of pid %d, error: %s", p.Pid, err.Error())
	}
	p.Exe = exe

	cmdline, err := psutilProc.Cmdline()
	if err != nil {
		return fmt.Errorf("Cannot load process cmdline of pid %d, error: %s", p.Pid, err.Error())
	}
	p.Cmdline = cmdline

	numThreads, err := psutilProc.NumThreads()
	if err != nil {
		return fmt.Errorf("Cannot load process num threads of pid %d, error: %s", p.Pid, err.Error())
	}
	p.NumThreads = int(numThreads)

	for _, child := range p.Children {
		if err := child.loadTreeDetails(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Process) addWmicChildren(children []*process.Process) {
	for _, child := range children {
		wrappedChild := &Process{Process: child.Process}
		wrappedChild.addWmicChildren(child.Children)
		p.Children = append(p.Children, wrappedChild)
	}
}

func (p *Process) findAndAddGoPsUtilChildren(children []*process2.Process) error {
	for _, child := range children {
		osChildProc, err := os.FindProcess(int(child.Pid))
		if err != nil {
			return fmt.Errorf("Unable to load child process with pid %d (parent pid %d), error: %s", child.Pid, p.Process.Pid, err.Error())
		}
		wrappedChild := &Process{Process: osChildProc}

		children, err := child.Children()
		if err != nil {
			return fmt.Errorf("Cannot load process (pid %d) children, error: %s", child.Pid, err.Error())
		}
		if err = wrappedChild.findAndAddGoPsUtilChildren(children); err != nil {
			return err
		}
		p.Children = append(p.Children, wrappedChild)
	}
	return nil
}

func (p *Process) flattenedPids() (pids []int) {
	pids = []int{
		p.Pid,
	}
	for _, child := range p.Children {
		pids = append(pids, child.flattenedPids()...)
	}
	return
}
