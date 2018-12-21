package process

// Lots of inspiration from: https://github.com/mitchellh/go-ps/blob/master/process_windows.go

import (
	"github.com/pkg/errors"
	"syscall"
	"unsafe"
)

// Windows API functions
var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
)

// Some constants from the Windows API
const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
)

// PROCESSENTRY32 is the Windows API structure that contains a process's
// information.
type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

// WindowsProcess is an implementation of Process for Windows.
type windowsProcess struct {
	pid  int
	ppid int
	exe  string
}

type windowsProcessInfo struct {
}

func (p *windowsProcess) Pid() int {
	return p.pid
}

func (p *windowsProcess) Name() (name string, harderror error, softerrors []error) {
	return p.exe, nil, nil
}

func (p *windowsProcess) Info() (ProcessInfo, error) {
	return &windowsProcessInfo{}, nil
}

func processFromPid(pid int) (Process, error, []error) {
	ps, harderr, softerrors := getAllProcesses()
	if harderr != nil {
		return nil, errors.Wrapf(harderr, "Unable to get all Pids"), softerrors
	}

	for _, p := range ps {
		if p.Pid() == pid {
			return p, nil, nil
		}
	}

	return nil, errors.Errorf("Unable to find process with pid %d", pid), softerrors
}

func getAllProcesses() (pids []Process, harderror error, softerrors []error) {
	handle, _, _ := procCreateToolhelp32Snapshot.Call(
		0x00000002,
		0)
	if handle < 0 {
		return nil, errors.Wrapf(syscall.GetLastError(), "Unable to get Windows processes."), nil
	}
	defer procCloseHandle.Call(handle)

	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, errors.New("Error retrieving process info."), nil
	}

	results := make([]Process, 0, 50)
	for {
		results = append(results, newWindowsProcess(&entry))

		ret, _, _ := procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return results, nil, nil
}

func newWindowsProcess(e *PROCESSENTRY32) *windowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return &windowsProcess{
		pid:  int(e.ProcessID),
		ppid: int(e.ParentProcessID),
		exe:  syscall.UTF16ToString(e.ExeFile[:end]),
	}
}
