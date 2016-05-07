package process_tree

import (
	"fmt"
	"os"

	"github.com/go-zero-boilerplate/wmic-proccess-tree/process"
	process2 "github.com/shirou/gopsutil/process"
)

type visitorGetTree struct {
	mainProcID int
	tree       *ProcessTree
	err        error
}

func (v *visitorGetTree) VisitWindows() {
	p, err := process.LoadProcessTree(v.mainProcID)
	if err != nil {
		v.err = fmt.Errorf("Cannot load process tree (pid %d), error: %s", v.mainProcID, err.Error())
		return
	}

	mainProc := &Process{Process: p.Process}
	mainProc.addWmicChildren(p.Children)

	if err = mainProc.loadTreeDetails(); err != nil {
		v.err = err
		return
	}
	v.tree = &ProcessTree{MainProcess: mainProc}
}

func (v *visitorGetTree) VisitLinux() {
	p, err := process2.NewProcess(int32(v.mainProcID))
	if err != nil {
		v.err = err
		return
	}

	children, err := p.Children()
	if err != nil {
		v.err = fmt.Errorf("Cannot load process (pid %d) children, error: %s", p.Pid, err.Error())
		return
	}

	osMainProc, err := os.FindProcess(int(p.Pid))
	if err != nil {
		v.err = fmt.Errorf("Cannot find os process by pid %d, error: %s", p.Pid, err.Error())
		return
	}

	mainProc := &Process{Process: osMainProc}
	if err = mainProc.findAndAddGoPsUtilChildren(children); err != nil {
		v.err = fmt.Errorf("Cannot add process (pid %d) children, error: %s", p.Pid, err.Error())
		return
	}

	if err = mainProc.loadTreeDetails(); err != nil {
		v.err = err
		return
	}
	v.tree = &ProcessTree{MainProcess: mainProc}
}

func (v *visitorGetTree) VisitDarwin() {
	v.VisitLinux() //GoPsUtils support both linux and Darwin but not Windows
}
