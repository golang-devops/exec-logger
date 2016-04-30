package main

//TODO: Write tests and strip this code into a separate repo/package - github.com/go-os-visitors/kill_process_tree

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/process"

	"github.com/go-zero-boilerplate/osvisitors"
)

func KillProcessTree(pid int, force bool) error {
	runtimeOsType, err := osvisitors.GetRuntimeOsType()
	if err != nil {
		return fmt.Errorf("Cannot get runtime OsType, error: %s", err.Error())
	}

	v := &killTreeOsVisitor{pid: pid, force: force}
	runtimeOsType.Accept(v)
	return v.err
}

type killTreeOsVisitor struct {
	pid   int
	force bool
	err   error
}

func (k *killTreeOsVisitor) VisitWindows() {
	cmdLine := []string{
		"TASKKILL",
		"/PID",
		fmt.Sprintf("%d", k.pid),
		"/T",
	}
	if k.force {
		cmdLine = append(cmdLine, "/F")
	}
	out, err := exec.Command(cmdLine[0], cmdLine[1:]...).CombinedOutput()
	if err != nil {
		k.err = fmt.Errorf("Cannot call TASKKILL. Error: %s. Output: %s", err.Error(), string(out))
		return
	}
}

func (k *killTreeOsVisitor) VisitLinux() {
	proc, err := process.NewProcess(int32(k.pid))
	if err != nil {
		k.err = fmt.Errorf("Unable to create NewProcess from pid %d, error: %s", k.pid, err.Error())
		return
	}

	errorStrs := []string{}

	children, err := proc.Children()
	if err != nil {
		errorStrs = append(errorStrs, fmt.Sprintf("Unable to get children of pid %d, error: %s", k.pid, err.Error()))
		// do not exit, we need to kill parent process still
	} else {
		for _, child := range children {
			errKillChild := child.Kill()
			if errKillChild != nil {
				errorStrs = append(errorStrs, fmt.Sprintf("Could not kill child process (pid %d) of parent process (pid %d), error: %s", child.Pid, k.pid, errKillChild.Error()))
			}
		}
	}

	err = proc.Kill()
	if err != nil {
		errorStrs = append(errorStrs, fmt.Sprintf("Unable to kill main process pid %d, error: %s", k.pid, err.Error()))
	}

	if len(errorStrs) > 0 {
		k.err = fmt.Errorf("Combined %d errors in attempt to kill process pid %d with its children: %s", len(errorStrs), k.pid, strings.Join(errorStrs, "\\n"))
		return
	}
}

func (k *killTreeOsVisitor) VisitDarwin() {
	k.VisitLinux()
}
